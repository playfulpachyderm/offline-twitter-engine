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

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

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
		UserID:          UserID(1423),
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
