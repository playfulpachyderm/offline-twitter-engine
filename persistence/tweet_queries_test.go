package persistence_test

import (
    "testing"
    "time"

    "github.com/go-test/deep"
)


/**
 * Create a Tweet, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadTweet(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()
    tweet.IsContentDownloaded = true

    // Save the tweet
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    // Reload the tweet
    new_tweet, err := profile.GetTweetById(tweet.ID)
    if err != nil {
        t.Fatalf("Failed to load the tweet: %s", err.Error())
    }

    if diff := deep.Equal(tweet, new_tweet); diff != nil {
        t.Error(diff)
    }
}

/**
 * Same as above, but with a tombstone
 */
func TestSaveAndLoadTombstone(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tombstone()

    // Save the tweet
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    // Reload the tweet
    new_tweet, err := profile.GetTweetById(tweet.ID)
    if err != nil {
        t.Fatalf("Failed to load the tweet: %s", err.Error())
    }

    if diff := deep.Equal(tweet, new_tweet); diff != nil {
        t.Error(diff)
    }
}

/**
 * Saving a tweet that already exists shouldn't reduce its backed-up status.
 * i.e., content which is already saved shouldn't be marked un-saved if it's removed from Twitter.
 * After all, that's the whole point of archiving.
 *
 * - is_stub should only go from "yes" to "no"
 * - is_content_downloaded should only go from "no" to "yes"
 */
func TestNoWorseningTweet(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()
    tweet.IsContentDownloaded = true
    tweet.IsStub = false
    tweet.IsConversationScraped = true
    tweet.LastScrapedAt = time.Unix(1000, 0)

    // Save the tweet
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    // Worsen the tweet and re-save it
    tweet.IsContentDownloaded = false
    tweet.IsStub = true
    tweet.IsConversationScraped = false
    tweet.LastScrapedAt = time.Unix(500, 0)
    err = profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    // Reload the tweet
    new_tweet, err := profile.GetTweetById(tweet.ID)
    if err != nil {
        t.Fatalf("Failed to load the tweet: %s", err.Error())
    }

    if new_tweet.IsStub != false {
        t.Errorf("Should have preserved non-stub status")
    }
    if new_tweet.IsContentDownloaded != true {
        t.Errorf("Should have preserved is-content-downloaded status")
    }
    if new_tweet.IsConversationScraped == false {
        t.Errorf("Should have preserved is-conversation-scraped status")
    }
    if new_tweet.LastScrapedAt.Unix() != 1000 {
        t.Errorf("Should have preserved last-scraped-at time")
    }
}

func TestModifyTweet(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()
    tweet.NumLikes = 1000
    tweet.NumRetweets = 2000
    tweet.NumReplies = 3000
    tweet.NumQuoteTweets = 4000
    tweet.IsStub = true
    tweet.IsContentDownloaded = false
    tweet.IsConversationScraped = false
    tweet.LastScrapedAt = time.Unix(1000, 0)

    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    tweet.NumLikes = 1500
    tweet.NumRetweets = 2500
    tweet.NumReplies = 3500
    tweet.NumQuoteTweets = 4500
    tweet.IsStub = false
    tweet.IsContentDownloaded = true
    tweet.IsConversationScraped = true
    tweet.LastScrapedAt = time.Unix(2000, 0)

    err = profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to re-save the tweet: %s", err.Error())
    }

    // Reload the tweet
    new_tweet, err := profile.GetTweetById(tweet.ID)
    if err != nil {
        t.Fatalf("Failed to load the tweet: %s", err.Error())
    }

    if new_tweet.NumLikes != 1500 {
        t.Errorf("Expected %d likes, got %d", 1500, new_tweet.NumLikes)
    }
    if new_tweet.NumRetweets != 2500 {
        t.Errorf("Expected %d retweets, got %d", 2500, new_tweet.NumRetweets)
    }
    if new_tweet.NumReplies != 3500 {
        t.Errorf("Expected %d replies, got %d", 1500, new_tweet.NumReplies)
    }
    if new_tweet.NumQuoteTweets != 4500 {
        t.Errorf("Expected %d quote tweets, got %d", 4500, new_tweet.NumQuoteTweets)
    }
    if new_tweet.IsStub != false {
        t.Errorf("Expected tweet to not be a stub, but it was")
    }
    if new_tweet.IsContentDownloaded != true {
        t.Errorf("Expected tweet content to be downloaded, but it wasn't")
    }
    if new_tweet.IsConversationScraped != true {
        t.Errorf("Expected conversation to be scraped, but it wasn't")
    }
    if new_tweet.LastScrapedAt.Unix() != 2000 {
        t.Errorf("Expected tweet to be scraped at %d (unix timestamp), but got %d", 2000, new_tweet.LastScrapedAt.Unix())
    }
}

/**
 * Should correctly report whether the User exists in the database
 */
func TestIsTweetInDatabase(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()

    exists := profile.IsTweetInDatabase(tweet.ID)
    if exists {
        t.Errorf("It shouldn't exist, but it does: %d", tweet.ID)
    }
    err := profile.SaveTweet(tweet)
    if err != nil {
        panic(err)
    }
    exists = profile.IsTweetInDatabase(tweet.ID)
    if !exists {
        t.Errorf("It should exist, but it doesn't: %d", tweet.ID)
    }
}

/**
 * Should correctly populate the `User` field on a Tweet
 */
func TestLoadUserForTweet(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()

    // Save the tweet
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Errorf("Failed to save the tweet: %s", err.Error())
    }


    if tweet.User != nil {
        t.Errorf("`User` field is already there for some reason: %v", tweet.User)
    }

    err = profile.LoadUserFor(&tweet)
    if err != nil {
        t.Errorf("Failed to load user: %s", err.Error())
    }

    if tweet.User == nil {
        t.Errorf("Did not load a user.  It is still nil.")
    }
}
