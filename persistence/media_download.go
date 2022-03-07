package persistence

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"offline_twitter/scraper"
)

type MediaDownloader interface {
	Curl(url string, outpath string) error
}

type DefaultDownloader struct{}

/**
 * Download a file over HTTP and save it.
 *
 * args:
 * - url: the remote file to download
 * - outpath: the path on disk to save it to
 */
func (d DefaultDownloader) Curl(url string, outpath string) error {
	println(url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error %s: %s", url, resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error downloading image %s:\n  %w", url, err)
	}

	err = os.WriteFile(outpath, data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing to path %s, url %s:\n  %w", outpath, url, err)
	}
	return nil
}

/**
 * Downloads an Image, and if successful, marks it as downloaded in the DB
 */
func (p Profile) download_tweet_image(img *scraper.Image, downloader MediaDownloader) error {
	outfile := path.Join(p.ProfileDir, "images", img.LocalFilename)
	err := downloader.Curl(img.RemoteURL, outfile)
	if err != nil {
		return err
	}
	img.IsDownloaded = true
	return p.SaveImage(*img)
}

/**
 * Downloads a Video and its thumbnail, and if successful, marks it as downloaded in the DB
 */
func (p Profile) download_tweet_video(v *scraper.Video, downloader MediaDownloader) error {
	// Download the video
	outfile := path.Join(p.ProfileDir, "videos", v.LocalFilename)
	err := downloader.Curl(v.RemoteURL, outfile)
	if err != nil {
		return err
	}

	// Download the thumbnail
	outfile = path.Join(p.ProfileDir, "video_thumbnails", v.ThumbnailLocalPath)
	err = downloader.Curl(v.ThumbnailRemoteUrl, outfile)
	if err != nil {
		return err
	}

	v.IsDownloaded = true
	return p.SaveVideo(*v)
}

/**
 * Downloads an URL thumbnail image, and if successful, marks it as downloaded in the DB
 */
func (p Profile) download_link_thumbnail(url *scraper.Url, downloader MediaDownloader) error {
	if url.HasCard && url.HasThumbnail {
		outfile := path.Join(p.ProfileDir, "link_preview_images", url.ThumbnailLocalPath)
		err := downloader.Curl(url.ThumbnailRemoteUrl, outfile)
		if err != nil {
			return err
		}
	}
	url.IsContentDownloaded = true
	return p.SaveUrl(*url)
}

/**
 * Download a tweet's video and picture content.
 *
 * Wraps the `DownloadTweetContentWithInjector` method with the default (i.e., real) downloader.
 */
func (p Profile) DownloadTweetContentFor(t *scraper.Tweet) error {
	return p.DownloadTweetContentWithInjector(t, DefaultDownloader{})
}

/**
 * Enable injecting a custom MediaDownloader (i.e., for testing)
 */
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

/**
 * Download a user's banner and profile images
 */
func (p Profile) DownloadUserContentFor(u *scraper.User) error {
	return p.DownloadUserContentWithInjector(u, DefaultDownloader{})
}

/**
 * Enable injecting a custom MediaDownloader (i.e., for testing)
 */
func (p Profile) DownloadUserContentWithInjector(u *scraper.User, downloader MediaDownloader) error {
	if !p.CheckUserContentDownloadNeeded(*u) {
		return nil
	}

	var outfile string
	var target_url string

	if u.ProfileImageUrl == "" {
		outfile = path.Join(p.ProfileDir, "profile_images", path.Base(scraper.DEFAULT_PROFILE_IMAGE_URL))
		target_url = scraper.DEFAULT_PROFILE_IMAGE_URL
	} else {
		outfile = path.Join(p.ProfileDir, "profile_images", u.ProfileImageLocalPath)
		target_url = u.ProfileImageUrl
	}

	err := downloader.Curl(target_url, outfile)
	if err != nil {
		return err
	}

	// Skip it if there's no banner image
	if u.BannerImageLocalPath != "" {
		outfile = path.Join(p.ProfileDir, "profile_images", u.BannerImageLocalPath)
		err = downloader.Curl(u.BannerImageUrl, outfile)

		if err != nil && strings.Contains(err.Error(), "404 Not Found") {
			// Try adding "600x200".  Not sure why this does this but sometimes it does.
			err = downloader.Curl(u.BannerImageUrl+"/600x200", outfile)
		}
		if err != nil {
			return err
		}
	}

	u.IsContentDownloaded = true
	return p.SaveUser(u)
}

/**
 * Download a User's tiny profile image, if it hasn't been downloaded yet.
 * If it has been downloaded, do nothing.
 * If this user should have a big profile picture, defer to the regular `DownloadUserContentFor` method.
 */
func (p Profile) DownloadUserProfileImageTiny(u *scraper.User) error {
	if p.IsFollowing(u.Handle) {
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
