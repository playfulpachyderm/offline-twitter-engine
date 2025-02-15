package persistence_test

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestSaveAndLoadNotification(t *testing.T) {
	profile_path := "test_profiles/TestNotificationQuery"
	profile := create_or_load_profile(profile_path)

	// Save it
	n := create_dummy_notification()
	profile.SaveNotification(n)

	// Check it comes back the same
	new_n := profile.GetNotification(n.ID)
	if diff := deep.Equal(n, new_n); diff != nil {
		t.Error(diff)
	}
}

func TestGetUnreadNotificationsCount(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	unread_notifs_count := profile.GetUnreadNotificationsCount(UserID(1488963321701171204), 1724372973735)
	assert.Equal(2, unread_notifs_count)
}

// Ensure that setting / blanking the Action[Re]TweetIDs works correctly
func TestLikesNotificationWithBothTweetsAndRetweets(t *testing.T) {
	profile_path := "test_profiles/TestNotificationQuery"
	profile := create_or_load_profile(profile_path)

	// Create a "like" on a Tweet
	n := create_dummy_notification()
	n.Type = NOTIFICATION_TYPE_LIKE
	n.ActionTweetID = create_stable_tweet().ID
	n.TweetIDs = []TweetID{n.ActionTweetID} // Overwrite the `dummy` slice
	n.ActionRetweetID = TweetID(0)
	n.RetweetIDs = []TweetID{}
	profile.SaveNotification(n)

	// Check it comes back the same
	new_n := profile.GetNotification(n.ID)
	if diff := deep.Equal(n, new_n); diff != nil {
		t.Error(diff)
	}

	// Now the user "likes" a Retweet too
	n.ActionTweetID = TweetID(0)
	n.ActionRetweetID = create_stable_retweet().RetweetID
	n.RetweetIDs = append(n.RetweetIDs, n.ActionRetweetID)
	profile.SaveNotification(n)

	// Check it comes back the same
	new_n = profile.GetNotification(n.ID)
	if diff := deep.Equal(n, new_n); diff != nil {
		t.Error(diff)
	}

	// Now the user "likes" another Tweet
	new_tweet := create_dummy_tweet()
	profile.SaveTweet(new_tweet)
	n.ActionTweetID = new_tweet.ID
	n.ActionRetweetID = TweetID(0)
	n.TweetIDs = append(n.TweetIDs, new_tweet.ID)
	profile.SaveNotification(n)

	// Check it comes back the same
	new_n = profile.GetNotification(n.ID)
	if diff := deep.Equal(n, new_n); diff != nil {
		t.Error(diff)
	}
}
