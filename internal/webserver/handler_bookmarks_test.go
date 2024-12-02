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
	assert.Len(tweets, 2)

	// Double check pagination works properly
	resp = do_request_with_active_user(httptest.NewRequest("GET", "/bookmarks?cursor=1800452344077464795", nil))
	require.Equal(resp.StatusCode, 200)

	root, err = html.Parse(resp.Body)
	require.NoError(err)
	tweets = cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 1)
}
