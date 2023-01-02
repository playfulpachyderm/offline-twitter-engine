package scraper_test

import (
	"encoding/json"
	"testing"
	"time"

	"net/http"
	"net/http/cookiejar"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "offline_twitter/scraper"
)

// TODO authentication: this has to be removed and replaced with an integration test once the feature is stable-ish
func TestAuthentication(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	username := "offline_twatter"
	password := "S1pKIW#eRT016iA@OFcK"

	api := NewGuestSession()
	api.LogIn(username, password)

	assert.True(api.IsAuthenticated)
	assert.NotEqual(api.CSRFToken, "")
	assert.Equal(api.UserHandle, UserHandle("Offline_Twatter"))

	response, err := api.GetLikesFor(1458284524761075714, "")
	require.NoError(err)
	trove, err := response.ToTweetTrove()
	require.NoError(err)
	assert.True(len(trove.Tweets) > 0)
}

// An API object should serialize and then deserialize to give the same session state from before.
func TestJsonifyApi(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cookie_jar, err := cookiejar.New(nil)
	require.NoError(err)
	cookie_jar.SetCookies(&TWITTER_BASE_URL, []*http.Cookie{
		{Name: "name1", Value: "name1", Secure: true},
		{Name: "name2", Value: "name2", HttpOnly: true},
	})
	api := API{
		UserHandle:      UserHandle("userhandle"),
		IsAuthenticated: true,
		GuestToken:      "guest token",
		Client: http.Client{
			Timeout: 10 * time.Second,
			Jar:     cookie_jar,
		},
		CSRFToken: "csrf token",
	}

	bytes, err := json.Marshal(api)
	require.NoError(err)
	var new_api API
	err = json.Unmarshal(bytes, &new_api)
	require.NoError(err)

	cookies := api.Client.Jar.Cookies(&TWITTER_BASE_URL)

	assert.Equal(cookies[0].Name, "name1")
	assert.Equal(cookies[0].Value, "name1")

	if diff := deep.Equal(api, new_api); diff != nil {
		t.Error(diff)
	}
}
