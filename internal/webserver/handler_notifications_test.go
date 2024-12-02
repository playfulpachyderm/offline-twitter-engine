package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestNotifications(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Notifications page
	req := httptest.NewRequest("GET", "/notifications", nil)
	resp := do_request_with_active_user(req)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 6)

	// Show more
	req = httptest.NewRequest("GET", "/notifications?cursor=1726604756351", nil)
	req.Header.Set("HX-Request", "true")
	resp = do_request_with_active_user(req)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 5)
}
