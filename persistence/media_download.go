package persistence

import (
    "fmt"
    "os"
    "path"
    "net/http"
    "io/ioutil"

    "offline_twitter/scraper"
)

type MediaDownloader interface {
    Curl(url string, outpath string) error
}

type DefaultDownloader struct {}

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
        return fmt.Errorf("Error downloading image %s: %s", url, err.Error())
    }

    err = os.WriteFile(outpath, data, 0644)
    if err != nil {
        return fmt.Errorf("Error writing to path: %s, url: %s: %s", outpath, url, err.Error())
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
 * Downloads an Video, and if successful, marks it as downloaded in the DB
 */
func (p Profile) download_tweet_video(v *scraper.Video, downloader MediaDownloader) error {
    outfile := path.Join(p.ProfileDir, "videos", v.LocalFilename)
    err := downloader.Curl(v.RemoteURL, outfile)
    if err != nil {
        return err
    }
    v.IsDownloaded = true
    return p.SaveVideo(*v)
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
    var err error
    var outfile string

    outfile = path.Join(p.ProfileDir, "profile_images", u.ProfileImageLocalPath)
    err = downloader.Curl(u.ProfileImageUrl, outfile)
    if err != nil {
        return err
    }

    // Skip it if there's no banner image
    if u.BannerImageLocalPath != "" {
        outfile = path.Join(p.ProfileDir, "profile_images", u.BannerImageLocalPath)
        err = downloader.Curl(u.BannerImageUrl, outfile)
        if err != nil {
            return err
        }
    }

    u.IsContentDownloaded = true
    return p.SaveUser(*u)
}
