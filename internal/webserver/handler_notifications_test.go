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

func TestNotifications(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Boilerplate for setting an active user
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.ActiveUser = scraper.User{ID: 1488963321701171204, Handle: "Offline_Twatter"} // Simulate a login

	// Notifications page
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notifications", nil)
	app.ServeHTTP(recorder, req)
	resp := recorder.Result()
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 6)

	// Show more
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/notifications?cursor=1726604756351", nil)
	req.Header.Set("HX-Request", "true")
	app.ServeHTTP(recorder, req)
	resp = recorder.Result()
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 5)
}
