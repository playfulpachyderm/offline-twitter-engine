package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// Some types to spy on a MediaDownloader
type SpyResult struct {
	url     string
	outpath string
}

// TODO: doesn't need to be a struct, can just be the spy; also, the pointer might be unnecessary
type FakeDownloader struct {
	Spy *[]SpyResult
}

func NewFakeDownloader() FakeDownloader {
	ret := FakeDownloader{}
	ret.Spy = &[]SpyResult{}
	return ret
}
func (d FakeDownloader) Curl(url string, outpath string) error {
	*d.Spy = append(*d.Spy, SpyResult{url, outpath})
	return nil
}
func (d FakeDownloader) Contains(result SpyResult) bool {
	for _, r := range *d.Spy {
		if r == result {
			return true
		}
	}
	return false
}

func test_all_downloaded(tweet Tweet, yes_or_no bool, t *testing.T) {
	error_msg := map[bool]string{
		true:  "Expected to be downloaded, but it wasn't",
		false: "Expected not to be downloaded, but it was",
	}[yes_or_no]

	assert.Len(t, tweet.Images, 2)
	assert.Len(t, tweet.Videos, 1)
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

// Downloading a Tweet's contents should mark the Tweet as downloaded
func TestDownloadTweetContent(t *testing.T) {
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_dummy_tweet()

	// Persist the tweet
	err := profile.SaveTweet(tweet)
	require.NoError(t, err)

	// Make sure everything is marked "not downloaded"
	test_all_downloaded(tweet, false, t)

	// Do the (fake) downloading
	err = profile.DownloadTweetContentWithInjector(&tweet, NewFakeDownloader())
	require.NoError(t, err)

	// It should all be marked "yes downloaded" now
	test_all_downloaded(tweet, true, t)

	// Reload the Tweet (check db); should also be "yes downloaded"
	new_tweet, err := profile.GetTweetById(tweet.ID)
	require.NoError(t, err)
	test_all_downloaded(new_tweet, true, t)
}

// Downloading a User's contents should mark the User as downloaded
func TestDownloadUserContent(t *testing.T) {
	assert := assert.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()

	// Persist the User
	err := profile.SaveUser(&user)
	require.NoError(t, err)

	// Make sure the User is marked "not downloaded"
	assert.False(user.IsContentDownloaded)

	// Do the (fake) downloading
	fake_downloader := NewFakeDownloader()
	err = profile.DownloadUserContentWithInjector(&user, fake_downloader)
	require.NoError(t, err)

	// Check that the downloader was called with the appropriate stuff
	assert.Len(*fake_downloader.Spy, 2)
	assert.True(fake_downloader.Contains(SpyResult{
		"profile image url",
		"test_profiles/TestMediaQueries/profile_images/profile image local path",
	}))
	assert.True(fake_downloader.Contains(SpyResult{
		"banner image url",
		"test_profiles/TestMediaQueries/profile_images/banner image local path",
	}))

	// The User should now be marked "yes downloaded"
	assert.True(user.IsContentDownloaded)

	// Reload the User (check db); should also be "yes downloaded"
	new_user, err := profile.GetUserByID(user.ID)
	require.NoError(t, err)
	assert.True(new_user.IsContentDownloaded)
}

// Should download the right stuff if User has no banner image and default profile image
func TestDownloadDefaultUserContent(t *testing.T) {
	assert := assert.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()
	user.BannerImageUrl = ""
	user.BannerImageLocalPath = ""
	user.ProfileImageUrl = ""

	// Do the (fake) downloading
	fake_downloader := NewFakeDownloader()
	err := profile.DownloadUserContentWithInjector(&user, fake_downloader)
	require.NoError(t, err)

	// Check that the downloader was called with the appropriate stuff
	assert.Len(*fake_downloader.Spy, 1)
	assert.True(fake_downloader.Contains(SpyResult{
		DEFAULT_PROFILE_IMAGE_URL,
		"test_profiles/TestMediaQueries/profile_images/default_profile.png",
	}))
}
