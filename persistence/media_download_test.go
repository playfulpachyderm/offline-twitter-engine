package persistence_test

import (
    "testing"

    "offline_twitter/scraper"
)

type FakeDownloader struct {}
func (d FakeDownloader) Curl(url string, outpath string) error { return nil }

func test_all_downloaded(tweet scraper.Tweet, yes_or_no bool, t *testing.T) {
    error_msg := map[bool]string{
        true: "Expected to be downloaded, but it wasn't",
        false: "Expected not to be downloaded, but it was",
    }[yes_or_no]

    if len(tweet.Images) != 2 {
        t.Errorf("Expected %d images, got %d", 2, len(tweet.Images))
    }
    if len(tweet.Videos) != 1 {
        t.Errorf("Expected %d videos, got %d", 1, len(tweet.Videos))
    }
    for _, img := range tweet.Images {
        if img.IsDownloaded != yes_or_no {
            t.Errorf("%s: ImageID %d", error_msg, img.ID)
        }
    }
    for _, vid := range tweet.Videos {
        if vid.IsDownloaded != yes_or_no {
            t.Errorf("Expected not to be downloaded, but it was: VideoID %d", vid.ID)
        }
    }
    if tweet.IsContentDownloaded != yes_or_no {
        t.Errorf("%s: the tweet", error_msg)
    }
}

/**
 * Create an Image, save it, reload it, and make sure it comes back the same
 */
func TestDownloadTweetContent(t *testing.T) {
    profile_path := "test_profiles/TestMediaQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()

    // Persist the tweet
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    // Make sure everything is marked "not downloaded"
    test_all_downloaded(tweet, false, t)

    // Do the (fake) downloading
    err = profile.DownloadTweetContentWithInjector(&tweet, FakeDownloader{})
    if err != nil {
        t.Fatalf("Error running fake download: %s", err.Error())
    }

    // It should all be marked "yes downloaded" now
    test_all_downloaded(tweet, true, t)

    // Reload the tweet (check db); should also be "yes downloaded"
    new_tweet, err := profile.GetTweetById(tweet.ID)
    if err != nil {
        t.Fatalf("Couldn't reload the tweet: %s", err.Error())
    }
    test_all_downloaded(new_tweet, true, t)
}
