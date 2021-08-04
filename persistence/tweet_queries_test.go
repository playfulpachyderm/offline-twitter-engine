package persistence_test

import (
    "testing"

    "github.com/go-test/deep"
)


/**
 * Create a Tweet, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadTweet(t *testing.T) {
    profile_path := "test_profiles/TestTweetQueries"
    profile := create_or_load_profile(profile_path)

    tweet := create_dummy_tweet()

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

    for i := range tweet.Videos {
        tweet.Videos[i].ID = new_tweet.Videos[i].ID
    }

    if diff := deep.Equal(tweet, new_tweet); diff != nil {
        t.Error(diff)
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
        t.Errorf("It shouldn't exist, but it does: %s", tweet.ID)
    }
    err := profile.SaveTweet(tweet)
    if err != nil {
        panic(err)
    }
    exists = profile.IsTweetInDatabase(tweet.ID)
    if !exists {
        t.Errorf("It should exist, but it doesn't: %s", tweet.ID)
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
