package persistence_test

import (
	"testing"

	"math/rand"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// Create an Image, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadImage(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()

	// Create a fresh Image to test on
	img := create_image_from_id(rand.Int())
	img.TweetID = tweet.ID

	// Save the Image
	err := profile.SaveImage(img)
	require.NoError(err)

	// Reload the Image
	imgs, err := profile.GetImagesForTweet(tweet)
	require.NoError(err)

	var new_img Image
	for index := range imgs {
		if imgs[index].ID == img.ID {
			new_img = imgs[index]
		}
	}
	require.Equal(img.ID, new_img.ID, "Could not find image for some reason")
	if diff := deep.Equal(img, new_img); diff != nil {
		t.Error(diff)
	}
}

// Change an Image, save the changes, reload it, and check if it comes back the same
func TestModifyImage(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()
	img := tweet.Images[0]

	require.Equal(ImageID(-1), img.ID, "Got the wrong image back")

	img.IsDownloaded = true

	// Save the changes
	err := profile.SaveImage(img)
	require.NoError(err)

	// Reload it
	imgs, err := profile.GetImagesForTweet(tweet)
	require.NoError(err)

	new_img := imgs[0]
	require.Equal(imgs[0], new_img, "Got the wrong image back")

	if diff := deep.Equal(img, new_img); diff != nil {
		t.Error(diff)
	}
}

// Create an Video, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadVideo(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()

	// Create a fresh Video to test on
	vid := create_video_from_id(rand.Int())
	vid.TweetID = tweet.ID
	vid.IsGif = true
	vid.IsBlockedByDMCA = true

	// Save the Video
	err := profile.SaveVideo(vid)
	require.NoError(err)

	// Reload the Video
	vids, err := profile.GetVideosForTweet(tweet)
	require.NoError(err)

	var new_vid Video
	for index := range vids {
		if vids[index].ID == vid.ID {
			new_vid = vids[index]
		}
	}
	require.Equal(vid.ID, new_vid.ID, "Could not find video for some reason")

	if diff := deep.Equal(vid, new_vid); diff != nil {
		t.Error(diff)
	}
}

// Change an Video, save the changes, reload it, and check if it comes back the same
func TestModifyVideo(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()
	vid := tweet.Videos[0]
	require.Equal(VideoID(-1), vid.ID, "Got the wrong video back")

	vid.IsDownloaded = true
	vid.IsBlockedByDMCA = true
	vid.ViewCount = 23000

	// Save the changes
	err := profile.SaveVideo(vid)
	require.NoError(err)

	// Reload it
	vids, err := profile.GetVideosForTweet(tweet)
	require.NoError(err)

	new_vid := vids[0]
	require.Equal(vid.ID, new_vid.ID, "Got the wrong video back")

	if diff := deep.Equal(vid, new_vid); diff != nil {
		t.Error(diff)
	}
}

// Create an Url, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadUrl(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()

	// Create a fresh Url to test on
	url := create_url_from_id(rand.Int())
	url.TweetID = tweet.ID

	// Save the Url
	err := profile.SaveUrl(url)
	require.NoError(err)

	// Reload the Url
	urls, err := profile.GetUrlsForTweet(tweet)
	require.NoError(err)

	var new_url Url
	for index := range urls {
		if urls[index].Text == url.Text {
			new_url = urls[index]
		}
	}
	require.Equal(url.Text, new_url.Text, "Could not find the url for some reason")

	if diff := deep.Equal(url, new_url); diff != nil {
		t.Error(diff)
	}
}

// Change an Url, save the changes, reload it, and check if it comes back the same
func TestModifyUrl(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()
	url := tweet.Urls[0]

	require.Equal("-1text", url.Text, "Got the wrong url back")

	url.IsContentDownloaded = true

	// Save the changes
	err := profile.SaveUrl(url)
	require.NoError(err)

	// Reload it
	urls, err := profile.GetUrlsForTweet(tweet)
	require.NoError(err)

	new_url := urls[0]
	require.Equal("-1text", url.Text, "Got the wrong url back")

	if diff := deep.Equal(url, new_url); diff != nil {
		t.Error(diff)
	}
}

// Create a Poll, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadPoll(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()

	poll := create_poll_from_id(rand.Int())
	poll.TweetID = tweet.ID

	// Save the Poll
	err := profile.SavePoll(poll)
	require.NoError(err)

	// Reload the Poll
	polls, err := profile.GetPollsForTweet(tweet)
	require.NoError(err)

	var new_poll Poll
	for index := range polls {
		if polls[index].ID == poll.ID {
			new_poll = polls[index]
		}
	}
	require.Equal(poll.ID, new_poll.ID, "Could not find poll for some reason")

	if diff := deep.Equal(poll, new_poll); diff != nil {
		t.Error(diff)
	}
}

// Change an Poll, save the changes, reload it, and check if it comes back the same
func TestModifyPoll(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	tweet := create_stable_tweet()
	poll := tweet.Polls[0]

	require.Equal("-1", poll.Choice1, "Got the wrong Poll back")

	poll.Choice1_Votes = 1200 // Increment it by 200 votes

	// Save the changes
	err := profile.SavePoll(poll)
	require.NoError(err)

	// Reload it
	polls, err := profile.GetPollsForTweet(tweet)
	require.NoError(err)

	new_poll := polls[0]
	require.Equal("-1", new_poll.Choice1, "Got the wrong poll back")

	if diff := deep.Equal(poll, new_poll); diff != nil {
		t.Error(diff)
	}
}
