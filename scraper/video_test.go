package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "offline_twitter/scraper"
)

func TestParseAPIVideo(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/tweet_content/video.json")
    if err != nil {
        panic(err)
    }
    var apivideo scraper.APIExtendedMedia
    err = json.Unmarshal(data, &apivideo)
    if err != nil {
        t.Fatal(err.Error())
    }
    tweet_id := scraper.TweetID(28)
    video := scraper.ParseAPIVideo(apivideo, tweet_id)

    expected_id := 1418951950020845568
    if video.ID != scraper.VideoID(expected_id) {
        t.Errorf("Expected ID of %d, got %d", expected_id, video.ID)
    }
    if video.TweetID != tweet_id {
        t.Errorf("Expected ID of %d, got %d", tweet_id, video.TweetID)
    }
    expected_remote_url := "https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12"
    if video.RemoteURL != expected_remote_url {
        t.Errorf("Expected %q, got %q", expected_remote_url, video.RemoteURL)
    }
    expected_height := 1280
    if video.Height != expected_height {
        t.Errorf("Expected height %d, got %d", expected_height, video.Height)
    }
    expected_width := 720
    if video.Width != expected_width {
        t.Errorf("Expected width %d, got %d", expected_width, video.Width)
    }

    expected_local_filename := "28.mp4"
    if video.LocalFilename != expected_local_filename {
        t.Errorf("Expected %q, got %q", expected_local_filename, video.LocalFilename)
    }
    if video.IsDownloaded {
        t.Errorf("Expected it not to be downloaded, but it was")
    }
}
