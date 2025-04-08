package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestNotificationsRequiresActiveSession(t *testing.T) {
	require := require.New(t)

	req := httptest.NewRequest("GET", "/notifications", nil)
	resp := do_request(req)
	require.Equal(401, resp.StatusCode)
}

func TestNotifications(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Notifications page
	req := httptest.NewRequest("GET", "/notifications", nil)
	resp := do_request_with_active_user(req)
	require.Equal(200, resp.StatusCode)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 6)
	// Check page title
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Notifications | Offline Twitter")

	// Show more
	req = httptest.NewRequest("GET", "/notifications?cursor=1726604756351", nil)
	req.Header.Set("HX-Request", "true")
	resp = do_request_with_active_user(req)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".notification")), 5)
}

// When scraping is disabled, marking notifs as read should 401
func TestNotificationsMarkAsRead(t *testing.T) {
	require := require.New(t)

	resp := do_request_with_active_user(httptest.NewRequest("GET", "/notifications/mark-all-as-read", nil))
	require.Equal(401, resp.StatusCode)
}
