package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestBookmarksTab(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Boilerplate for setting an active user
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.ActiveUser = scraper.User{ID: 1488963321701171204, Handle: "Offline_Twatter"} // Simulate a login

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest("GET", "/bookmarks", nil))
	resp := recorder.Result()
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 2)

	// Double check pagination works properly
	recorder = httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest("GET", "/bookmarks?cursor=1800452344077464795", nil))
	resp = recorder.Result()
	require.Equal(resp.StatusCode, 200)

	root, err = html.Parse(resp.Body)
	require.NoError(err)
	tweets = cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 1)
}
