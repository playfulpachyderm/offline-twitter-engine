package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestBookmarksTab(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request_with_active_user(httptest.NewRequest("GET", "/bookmarks", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Bookmarks | Offline Twitter")
	assert.Len(tweets, 2)

	// With pagination
	req := httptest.NewRequest("GET", "/bookmarks?cursor=1800452344077464795", nil)
	req.Header.Set("HX-Request", "true")
	resp = do_request_with_active_user(req)
	require.Equal(resp.StatusCode, 200)

	root, err = html.Parse(resp.Body)
	require.NoError(err)
	tweets = cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(tweets, 1)
}

// When scraping is disabled, should 401
func TestBookmarksScrape(t *testing.T) {
	require := require.New(t)

	// Attempt to scrape with scraping disabled
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/bookmarks?scrape", nil))
	require.Equal(resp.StatusCode, 401)
}

// If cursor is invalid, it should 400
func TestBookmarksInvalidCursor(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// HTMX version
	req := httptest.NewRequest("GET", "/bookmarks?cursor=asdf", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request_with_active_user(req)
	require.Equal(resp.StatusCode, 400)
	// Piggyback in testing of HTMX 400 error toasts
	assert.Equal("beforeend", resp.Header.Get("HX-Reswap"))
	assert.Equal("#toasts", resp.Header.Get("HX-Retarget"))
	assert.Equal("false", resp.Header.Get("HX-Push-Url"))

	// Non-HTMX version
	req1 := httptest.NewRequest("GET", "/bookmarks?cursor=asdf", nil)
	resp1 := do_request_with_active_user(req1)
	require.Equal(resp1.StatusCode, 400)
	assert.Equal("", resp1.Header.Get("HX-Reswap"))
	assert.Equal("", resp1.Header.Get("HX-Retarget"))
	assert.Equal("", resp1.Header.Get("HX-Push-Url"))
}
