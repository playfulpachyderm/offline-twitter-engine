package persistence_test

import (
    "testing"

    "offline_twitter/scraper"
)

/**
 * Should return the base name (filename) of the remote URL
 */
func TestImageFilenameWhenDownloaded(t *testing.T) {
    i := scraper.Image{IsDownloaded: false, Filename: "https://pbs.twimg.com/media/E7vG1kxWQAQrWGF.jpg"}
    outpath := i.FilenameWhenDownloaded()
    expected := "E7vG1kxWQAQrWGF.jpg"
    if outpath != expected {
        t.Errorf("Expected output path to be %q, but got %q", expected, outpath)
    }
}


/**
 * Should return an `.mp4`file matching its parent Tweet's ID
 */
func TestVideoFilenameWhenDownloaded(t *testing.T) {
    v := scraper.Video{TweetID: scraper.TweetID("23"), IsDownloaded: false, Filename: "https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/320x568/IXaQ5rPyf9mbD1aD.mp4?tag=12"}
    outpath := v.FilenameWhenDownloaded()
    expected := "23.mp4"
    if outpath != expected {
        t.Errorf("Expected output path to be %q, but got %q", expected, outpath)
    }
}
