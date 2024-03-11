package persistence

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type MediaDownloader interface {
	Curl(url string, outpath string) error
}

type DefaultDownloader struct{}

// Download a file over HTTP and save it.
//
// args:
// - url: the remote file to download
// - outpath: the path on disk to save it to
func (d DefaultDownloader) Curl(url string, outpath string) error {
	data, err := scraper.DownloadMedia(url)
	if err != nil {
		return fmt.Errorf("downloading %q:\n  %w", url, err)
	}

	// Ensure the output directory exists
	dirname := path.Dir(outpath)
	if dirname != "." {
		err = os.MkdirAll(dirname, 0755)
		if err != nil {
			panic(err)
		}
	}

	// Write the downloaded data
	err = os.WriteFile(outpath, data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing to path %s, url %s:\n  %w", outpath, url, err)
	}
	return nil
}

// Downloads an Image, and if successful, marks it as downloaded in the DB
func (p Profile) download_tweet_image(img *scraper.Image, downloader MediaDownloader) error {
	outfile := path.Join(p.ProfileDir, "images", img.LocalFilename)
	err := downloader.Curl(img.RemoteURL, outfile)
	if err != nil {
		return fmt.Errorf("Error downloading tweet image (TweetID %d):\n  %w", img.TweetID, err)
	}
	img.IsDownloaded = true
	return p.SaveImage(*img)
}

// Downloads a Video and its thumbnail, and if successful, marks it as downloaded in the DB
func (p Profile) download_tweet_video(v *scraper.Video, downloader MediaDownloader) error {
	// Download the video
	outfile := path.Join(p.ProfileDir, "videos", v.LocalFilename)
	err := downloader.Curl(v.RemoteURL, outfile)

	if errors.Is(err, scraper.ErrorDMCA) {
		v.IsDownloaded = false
		v.IsBlockedByDMCA = true
	} else if err != nil {
		return fmt.Errorf("Error downloading video (TweetID %d):\n  %w", v.TweetID, err)
	} else {
		v.IsDownloaded = true
	}

	// Download the thumbnail
	outfile = path.Join(p.ProfileDir, "video_thumbnails", v.ThumbnailLocalPath)
	err = downloader.Curl(v.ThumbnailRemoteUrl, outfile)
	if err != nil {
		v.IsDownloaded = false
		return fmt.Errorf("Error downloading video thumbnail (TweetID %d):\n  %w", v.TweetID, err)
	}

	return p.SaveVideo(*v)
}

// Downloads an URL thumbnail image, and if successful, marks it as downloaded in the DB
func (p Profile) download_link_thumbnail(url *scraper.Url, downloader MediaDownloader) error {
	if url.HasCard && url.HasThumbnail {
		outfile := path.Join(p.ProfileDir, "link_preview_images", url.ThumbnailLocalPath)
		err := downloader.Curl(url.ThumbnailRemoteUrl, outfile)
		if err != nil {
			return fmt.Errorf("Error downloading link thumbnail (TweetID %d):\n  %w", url.TweetID, err)
		}
	}
	url.IsContentDownloaded = true
	return p.SaveUrl(*url)
}

// Download a tweet's video and picture content.
// Wraps the `DownloadTweetContentWithInjector` method with the default (i.e., real) downloader.
func (p Profile) DownloadTweetContentFor(t *scraper.Tweet) error {
	return p.DownloadTweetContentWithInjector(t, DefaultDownloader{})
}

// Enable injecting a custom MediaDownloader (i.e., for testing)
func (p Profile) DownloadTweetContentWithInjector(t *scraper.Tweet, downloader MediaDownloader) error {
	// Check if content needs to be downloaded; if not, just return
	if !p.CheckTweetContentDownloadNeeded(*t) {
		return nil
	}

	for i := range t.Images {
		err := p.download_tweet_image(&t.Images[i], downloader)
		if err != nil {
			return err
		}
	}

	for i := range t.Videos {
		// Videos can be geoblocked, and the HTTP response isn't in JSON so it's hard to capture
		if t.Videos[i].IsGeoblocked {
			continue
		}

		err := p.download_tweet_video(&t.Videos[i], downloader)
		if err != nil {
			return err
		}
	}

	for i := range t.Urls {
		err := p.download_link_thumbnail(&t.Urls[i], downloader)
		if err != nil {
			return err
		}
	}
	t.IsContentDownloaded = true
	return p.SaveTweet(*t)
}

// Download a user's banner and profile images
func (p Profile) DownloadUserContentFor(u *scraper.User) error {
	return p.DownloadUserContentWithInjector(u, DefaultDownloader{})
}

// Enable injecting a custom MediaDownloader (i.e., for testing)
func (p Profile) DownloadUserContentWithInjector(u *scraper.User, downloader MediaDownloader) error {
	if !p.CheckUserContentDownloadNeeded(*u) {
		return nil
	}

	outfile := p.get_profile_image_output_path(*u)

	var target_url string
	if u.ProfileImageUrl == "" {
		target_url = scraper.DEFAULT_PROFILE_IMAGE_URL
	} else {
		target_url = u.ProfileImageUrl
	}

	err := downloader.Curl(target_url, outfile)
	if err != nil {
		return fmt.Errorf("Error downloading profile image for user %q:\n  %w", u.Handle, err)
	}

	// Skip it if there's no banner image
	if u.BannerImageLocalPath != "" {
		outfile = p.get_banner_image_output_path(*u)
		err = downloader.Curl(u.BannerImageUrl, outfile)

		if err != nil && strings.Contains(err.Error(), "404 Not Found") {
			// Try adding "600x200".  Not sure why this does this but sometimes it does.
			err = downloader.Curl(u.BannerImageUrl+"/600x200", outfile)
		}
		if err != nil {
			return fmt.Errorf("Error downloading banner image for user %q:\n  %w", u.Handle, err)
		}
	}

	u.IsContentDownloaded = true
	return p.SaveUser(u)
}

// Download a User's tiny profile image, if it hasn't been downloaded yet.
// If it has been downloaded, do nothing.
// If this user should have a big profile picture, defer to the regular `DownloadUserContentFor` method.
func (p Profile) DownloadUserProfileImageTiny(u *scraper.User) error {
	if p.IsFollowing(*u) {
		return p.DownloadUserContentFor(u)
	}

	d := DefaultDownloader{}

	outfile := path.Join(p.ProfileDir, "profile_images", u.GetTinyProfileImageLocalPath())
	if file_exists(outfile) {
		return nil
	}
	err := d.Curl(u.GetTinyProfileImageUrl(), outfile)
	return err
}
