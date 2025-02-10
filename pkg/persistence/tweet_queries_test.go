package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Create a Tweet, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadTweet(t *testing.T) {
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.IsContentDownloaded = true

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(t, err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(t, err)

	if diff := deep.Equal(tweet, new_tweet); diff != nil {
		t.Error(diff)
	}
}

// Same as above, but with a tombstone
func TestSaveAndLoadTombstone(t *testing.T) {
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tombstone()

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(t, err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(t, err)

	if diff := deep.Equal(tweet, new_tweet); diff != nil {
		t.Error(diff)
	}
}

// Saving a tweet that already exists shouldn't reduce its backed-up status.
// i.e., content which is already saved shouldn't be marked un-saved if it's removed from Twitter.
// After all, that's the whole point of archiving.
//
// - is_stub should only go from "yes" to "no"
// - is_content_downloaded should only go from "no" to "yes"
func TestNoWorseningTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.IsContentDownloaded = true
	tweet.IsStub = false
	tweet.IsConversationScraped = true
	tweet.IsExpandable = true
	tweet.LastScrapedAt = TimestampFromUnix(1000)
	tweet.Text = "Yes text"
	tweet.NumLikes = 10
	tweet.NumRetweets = 11
	tweet.NumQuoteTweets = 12
	tweet.NumReplies = 13

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(err)

	// Worsen the tweet and re-save it
	tweet.IsContentDownloaded = false
	tweet.IsStub = true
	tweet.IsConversationScraped = false
	tweet.IsExpandable = false
	tweet.LastScrapedAt = TimestampFromUnix(500)
	tweet.Text = ""
	err = profile.SaveTweet(tweet)
	require.NoError(err)
	tweet.NumLikes = 0
	tweet.NumRetweets = 0
	tweet.NumQuoteTweets = 0
	tweet.NumReplies = 0

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(err)

	assert.False(new_tweet.IsStub, "Should have preserved non-stub status")
	assert.True(new_tweet.IsContentDownloaded, "Should have preserved is-content-downloaded status")
	assert.True(new_tweet.IsConversationScraped, "Should have preserved is-conversation-scraped status")
	assert.True(new_tweet.IsExpandable)
	assert.Equal(int64(1000), new_tweet.LastScrapedAt.Unix(), "Should have preserved last-scraped-at time")
	assert.Equal(new_tweet.Text, "Yes text", "Text should not get clobbered if it becomes unavailable")
	assert.Equal(10, new_tweet.NumLikes)
	assert.Equal(11, new_tweet.NumRetweets)
	assert.Equal(12, new_tweet.NumQuoteTweets)
	assert.Equal(13, new_tweet.NumReplies)
}

// The tweet was a tombstone and is now available; it should be updated
func TestUntombstoningTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.TombstoneType = "hidden" // e.g., account was priv
	tweet.IsStub = true
	tweet.Text = ""

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(err)

	// Tweet suddenly becomes available
	tweet.TombstoneType = ""
	tweet.IsStub = false
	tweet.Text = "Some text"
	err = profile.SaveTweet(tweet)
	require.NoError(err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(err)

	assert.False(new_tweet.IsStub, "Should no longer be a stub after re-scrape")
	assert.Equal(new_tweet.TombstoneType, "", "Tweet shouldn't be a tombstone anymore")
	assert.Equal(new_tweet.Text, "Some text", "Should have created the text")
}

// The tweet is an expanding tweet, but was saved before expanding tweets were implemented
func TestUpgradingExpandingTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.IsExpandable = false
	tweet.Text = "Some long but cut-off text..."

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(err)

	// Now that we have expanding tweets
	tweet.IsExpandable = true
	tweet.Text = "Some long but cut-off text, but now it no longer is cut off!"
	err = profile.SaveTweet(tweet)
	require.NoError(err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(err)

	assert.True(new_tweet.IsExpandable, "Should now be is_expanding after re-scrape")
	assert.Equal(new_tweet.Text, "Some long but cut-off text, but now it no longer is cut off!", "Should have extended the text")
}

// The "unavailable" tombstone type is not reliable, you should be able to update away from it but
// not toward it
func TestChangingTombstoningTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.TombstoneType = "unavailable"
	tweet.IsStub = true
	tweet.Text = ""

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(err)

	// New tombstone type
	tweet.TombstoneType = "hidden"
	err = profile.SaveTweet(tweet)
	require.NoError(err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(err)

	assert.Equal(new_tweet.TombstoneType, "hidden", "Should be able to overwrite 'unavailable' tombstone")

	// New tombstone type
	new_tweet.TombstoneType = "hidden"
	err = profile.SaveTweet(new_tweet)
	require.NoError(err)

	// Reload the tweet
	new_tweet2, err := profile.GetTweetById(new_tweet.ID)
	require.NoError(err)

	assert.Equal(new_tweet2.TombstoneType, "hidden", "'Unavailable' shouldn't clobber other tombstone types")
}

func TestModifyTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

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
	tweet.LastScrapedAt = TimestampFromUnix(1000)

	err := profile.SaveTweet(tweet)
	require.NoError(err)

	tweet.NumLikes = 1500
	tweet.NumRetweets = 2500
	tweet.NumReplies = 3500
	tweet.NumQuoteTweets = 4500
	tweet.IsStub = false
	tweet.IsContentDownloaded = true
	tweet.IsConversationScraped = true
	tweet.LastScrapedAt = TimestampFromUnix(2000)
	tweet.TombstoneType = "deleted"

	err = profile.SaveTweet(tweet)
	require.NoError(err)

	// Reload the tweet
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(err)

	assert.Equal(1500, new_tweet.NumLikes)
	assert.Equal(2500, new_tweet.NumRetweets)
	assert.Equal(3500, new_tweet.NumReplies)
	assert.Equal(4500, new_tweet.NumQuoteTweets)
	assert.False(new_tweet.IsStub)
	assert.True(new_tweet.IsContentDownloaded)
	assert.True(new_tweet.IsConversationScraped)
	assert.Equal(int64(2000), new_tweet.LastScrapedAt.Unix())
	assert.Equal(new_tweet.TombstoneType, "deleted")
}

// Should correctly report whether the User exists in the database
func TestIsTweetInDatabase(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()

	exists := profile.IsTweetInDatabase(tweet.ID)
	require.False(exists)

	err := profile.SaveTweet(tweet)
	require.NoError(err)

	exists = profile.IsTweetInDatabase(tweet.ID)
	assert.True(t, exists)
}

// Should correctly populate the `User` field on a Tweet
func TestLoadUserForTweet(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(err)
	require.Nil(tweet.User, "`User` field is already there for some reason")

	err = profile.LoadUserFor(&tweet)
	require.NoError(err)
	require.NotNil(tweet.User, "Did not load a user.  It is still nil.")
}

// Test all the combinations for whether a tweet needs its content downloaded
func TestCheckTweetContentDownloadNeeded(t *testing.T) {
	assert := assert.New(t)
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()
	tweet.IsContentDownloaded = false

	// Non-saved tweets should need to be downloaded
	assert.True(profile.CheckTweetContentDownloadNeeded(tweet))

	// Save the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(t, err)

	// Should still need a download since `is_content_downloaded` is false
	assert.True(profile.CheckTweetContentDownloadNeeded(tweet))

	// Try again but this time with `is_content_downloaded` = true
	tweet.IsContentDownloaded = true
	err = profile.SaveTweet(tweet)
	require.NoError(t, err)

	// Should no longer need a download
	assert.False(profile.CheckTweetContentDownloadNeeded(tweet))
}

func TestLoadMissingTweet(t *testing.T) {
	profile_path := "test_profiles/TestTweetQueries"
	profile := create_or_load_profile(profile_path)

	_, err := profile.GetTweetById(TweetID(6234234)) // Random number
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotInDatabase)
}
