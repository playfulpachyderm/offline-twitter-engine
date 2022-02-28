package scraper_test

import (
	"testing"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jarcoal/httpmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

	. "offline_twitter/scraper"
)

func TestParseSingleUser(t *testing.T) {
	assert := assert.New(t)
	data, err := ioutil.ReadFile("test_responses/michael_malice_user_profile.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(t, err)

	apiUser := user_resp.ConvertToAPIUser()

	user, err := ParseSingleUser(apiUser)
	require.NoError(t, err)

	assert.Equal(UserID(44067298), user.ID)
	assert.Equal("Michael Malice", user.DisplayName)
	assert.Equal(UserHandle("michaelmalice"), user.Handle)
	assert.Equal("Author of Dear Reader, The New Right & The Anarchist Handbook\nHost of \"YOUR WELCOME\" \nSubject of Ego & Hubris by Harvey Pekar\nUnderwear Model\nHe/Him ⚑", user.Bio)
	assert.Equal(941, user.FollowingCount)
	assert.Equal(208589, user.FollowersCount)
	assert.Equal("Brooklyn", user.Location)
	assert.Equal("https://amzn.to/3oInafv", user.Website)
	assert.Equal(int64(1243920952), user.JoinDate.Unix())
	assert.False(user.IsPrivate)
	assert.True (user.IsVerified)
	assert.False(user.IsBanned)
	assert.Equal("https://pbs.twimg.com/profile_images/1064051934812913664/Lbwdb_C9.jpg", user.ProfileImageUrl)
	assert.Equal("https://pbs.twimg.com/profile_images/1064051934812913664/Lbwdb_C9_normal.jpg", user.GetTinyProfileImageUrl())
	assert.Equal("https://pbs.twimg.com/profile_banners/44067298/1615134676", user.BannerImageUrl)
	assert.Equal("michaelmalice_profile_Lbwdb_C9.jpg", user.ProfileImageLocalPath)
	assert.Equal("michaelmalice_banner_1615134676.jpg", user.BannerImageLocalPath)
	assert.Equal(TweetID(1403835414373339136), user.PinnedTweetID)
}

/**
 * Should correctly parse a banned user
 */
func TestParseBannedUser(t *testing.T) {
	assert := assert.New(t)
	data, err := ioutil.ReadFile("test_responses/suspended_user.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(t, err)

	apiUser := user_resp.ConvertToAPIUser()

	user, err := ParseSingleUser(apiUser)
	require.NoError(t, err)
	assert.Equal(UserID(193918550), user.ID)
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
	data, err := ioutil.ReadFile("test_responses/deleted_user.json")
	if err != nil {
		panic(err)
	}
	var user_resp UserResponse
	err = json.Unmarshal(data, &user_resp)
	require.NoError(t, err)

	handle := "Some Random Deleted User"

	apiUser := user_resp.ConvertToAPIUser()
	apiUser.ScreenName = string(handle)  // This is done in scraper.GetUser, since users are retrieved by handle anyway

	user, err := ParseSingleUser(apiUser)
	require.NoError(t, err)
	assert.Equal(UserID(0), user.ID)
	assert.True(user.IsIdFake)
	assert.True(user.IsNeedingFakeID)
	assert.Equal(user.Bio, "<blank>")
	assert.Equal(user.Handle, UserHandle(handle))

	// Test generation of profile images for deleted user
	assert.Equal("https://abs.twimg.com/sticky/default_profile_images/default_profile.png", user.GetTinyProfileImageUrl())
	assert.Equal("default_profile.png", user.GetTinyProfileImageLocalPath())
}

/**
 * Should extract a user handle from a tweet URL, or fail if URL is invalid
 */
func TestParseHandleFromTweetUrl(t *testing.T) {
	assert := assert.New(t)

	// Test valid tweet url
	url := "https://twitter.com/kanesays23/status/1429583672827465730"
	result, err := ParseHandleFromTweetUrl(url)
	assert.NoError(err)
	assert.Equal(UserHandle("kanesays23"), result)

	// Test url with GET params
	result, err = ParseHandleFromTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720?s=20")
	assert.NoError(err)
	assert.Equal(UserHandle("NerdNoticing"), result)

	// Test invalid url
	_, err = ParseHandleFromTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720s=20")
	assert.Error(err)

	// Test empty string
	_, err = ParseHandleFromTweetUrl("")
	assert.Error(err)
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
