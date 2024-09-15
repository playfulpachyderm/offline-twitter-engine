package scraper_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseSingleUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/michael_malice_user_profile.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(err)

	apiUser, err := user_resp.ConvertToAPIUser()
	require.NoError(err)

	user, err := ParseSingleUser(apiUser)
	require.NoError(err)

	assert.Equal(UserID(44067298), user.ID)
	assert.Equal("Michael Malice", user.DisplayName)
	assert.Equal(UserHandle("michaelmalice"), user.Handle)
	assert.Equal("Author: Dear Reader, The New Right, The Anarchist Handbook & The White Pill \n"+
		"Host: \"YOUR WELCOME\" \nSubject: Ego & Hubris by Harvey Pekar\nHe/Him ⚑", user.Bio)
	assert.Equal(1035, user.FollowingCount)
	assert.Equal(649484, user.FollowersCount)
	assert.Equal("Austin", user.Location)
	assert.Equal("https://amzn.to/3oInafv", user.Website)
	assert.Equal(int64(1243920952), user.JoinDate.Unix())
	assert.False(user.IsPrivate)
	assert.True(user.IsVerified)
	assert.False(user.IsBanned)
	assert.Equal("https://pbs.twimg.com/profile_images/1415820415314931715/_VVX4GI8.jpg", user.ProfileImageUrl)
	assert.Equal("https://pbs.twimg.com/profile_images/1415820415314931715/_VVX4GI8_normal.jpg", user.GetTinyProfileImageUrl())
	assert.Equal("https://pbs.twimg.com/profile_banners/44067298/1664774013", user.BannerImageUrl)
	assert.Equal("michaelmalice_profile__VVX4GI8.jpg", user.ProfileImageLocalPath)
	assert.Equal("michaelmalice_banner_1664774013.jpg", user.BannerImageLocalPath)
	assert.Equal(TweetID(1692611652397453790), user.PinnedTweetID)
}

/**
 * Should correctly parse a banned user
 */
func TestParseBannedUser(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/user_suspended.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(t, err)

	apiUser, err := user_resp.ConvertToAPIUser()
	require.NoError(t, err)

	user, err := ParseSingleUser(apiUser)
	require.NoError(t, err)
	assert.True(user.IsBanned)

	// Test generation of profile images for banned user
	assert.Equal("https://abs.twimg.com/sticky/default_profile_images/default_profile.png", user.GetTinyProfileImageUrl())
	assert.Equal("default_profile.png", user.GetTinyProfileImageLocalPath())
}

/**
 * Should correctly parse a deleted user
 */
func TestParseDeletedUser(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/deleted_user.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(t, err)

	_, err = user_resp.ConvertToAPIUser()
	assert.Error(err)
	assert.ErrorIs(err, ErrDoesntExist)
}

/**
 * Should extract a user handle from a shortened tweet URL
 */
func TestParseHandleFromShortenedTweetUrl(t *testing.T) {
	assert := assert.New(t)

	short_url := "https://t.co/rZVrNGJyDe"
	expanded_url := "https://twitter.com/MarkSnyderJr1/status/1460857606147350529"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", short_url, func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Location", expanded_url)
		return &http.Response{StatusCode: 301, Header: header}, nil
	})

	// Check the httpmock interceptor is working correctly
	require.Equal(t, expanded_url, ExpandShortUrl(short_url), "httpmock didn't intercept the request")

	result, err := ParseHandleFromTweetUrl(short_url)
	require.NoError(t, err)
	assert.Equal(UserHandle("MarkSnyderJr1"), result)
}

// Should compute tiny profile image URLs correctly, and fix local paths if needed (e.g., "_normal" and no file extension)
func TestGetTinyURLs(t *testing.T) {
	assert := assert.New(t)
	u := User{
		ProfileImageUrl: "https://pbs.twimg.com/profile_images/1208124284/iwRReicO.jpg",
		Handle:          "testUser",
	}
	assert.Equal(u.GetTinyProfileImageUrl(), "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal.jpg")
	assert.Equal(u.GetTinyProfileImageLocalPath(), "testUser_profile_iwRReicO_normal.jpg")

	// User with poorly formed profile image URL
	u.ProfileImageUrl = "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal"
	assert.Equal(u.GetTinyProfileImageUrl(), "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal")
	assert.Equal(u.GetTinyProfileImageLocalPath(), "testUser_profile_iwRReicO_normal.jpg")
}
