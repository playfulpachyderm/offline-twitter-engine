package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "offline_twitter/scraper"
)

func TestParseAPIMedia(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/tweet_content/image.json")
    if err != nil {
        panic(err)
    }
    var apimedia scraper.APIMedia
    err = json.Unmarshal(data, &apimedia)
    if err != nil {
        t.Fatal(err.Error())
    }
    image := scraper.ParseAPIMedia(apimedia)

    expected_id := 1395882862289772553
    if image.ID != scraper.ImageID(expected_id) {
        t.Errorf("Expected ID of %q, got %q", expected_id, image.ID)
    }
    expected_remote_url := "https://pbs.twimg.com/media/E18sEUrWYAk8dBl.jpg"
    if image.RemoteURL != expected_remote_url {
        t.Errorf("Expected %q, got %q", expected_remote_url, image.RemoteURL)
    }
    expected_local_filename := "E18sEUrWYAk8dBl.jpg"
    if image.LocalFilename != expected_local_filename {
        t.Errorf("Expected %q, got %q", expected_local_filename, image.LocalFilename)
    }
    if image.IsDownloaded {
        t.Errorf("Expected it not to be downloaded, but it was")
    }
}
