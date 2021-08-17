package persistence_test

import (
	"testing"

	"github.com/go-test/deep"
)


func TestSaveAndLoadRetweet(t *testing.T) {
	profile_path := "test_profiles/TestRetweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
    err := profile.SaveTweet(tweet)
    if err != nil {
        t.Fatalf("Failed to save the tweet: %s", err.Error())
    }

    rt := create_dummy_retweet(tweet.ID)

    // Save the Retweet
    err = profile.SaveRetweet(rt)
    if err != nil {
    	t.Fatalf("Failed to save the retweet: %s", err.Error())
    }

    // Reload the Retweet
    new_rt, err := profile.GetRetweetById(rt.RetweetID)
    if err != nil {
    	t.Fatalf("Failed to load the retweet: %s", err.Error())
    }

    if diff := deep.Equal(rt, new_rt); diff != nil {
    	t.Error(diff)
    }
}
