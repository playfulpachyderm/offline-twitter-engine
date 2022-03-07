package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"
)

func TestSaveAndLoadRetweet(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestRetweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	err := profile.SaveTweet(tweet)
	require.NoError(err)

	rt := create_dummy_retweet(tweet.ID)

	// Save the Retweet
	err = profile.SaveRetweet(rt)
	require.NoError(err)

	// Reload the Retweet
	new_rt, err := profile.GetRetweetById(rt.RetweetID)
	require.NoError(err)

	if diff := deep.Equal(rt, new_rt); diff != nil {
		t.Error(diff)
	}
}
