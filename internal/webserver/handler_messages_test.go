package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// Loading the index page should work if you're logged in
func TestMessagesIndexPage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat list
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/messages", nil))
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".chat-list .chat-list-entry")), 2)
	assert.Len(cascadia.QueryAll(root, selector(".chat-view .dm-message")), 0) // No messages until you click on one
}

// Open a chat room
func TestMessagesRoom(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat detail
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328", nil))
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".chat-list .chat-list-entry")), 2) // Chat list still renders
	assert.Len(cascadia.QueryAll(root, selector("#chat-view .dm-message")), 5)

	// Should have the poller at the bottom
	poller := cascadia.Query(root, selector("#new-messages-poller"))
	assert.NotNil(poller)
	assert.Contains(poller.Attr, html.Attribute{Key: "hx-get", Val: "/messages/1488963321701171204-1178839081222115328"})
	assert.Contains(
		cascadia.Query(poller, selector("input[name='scroll_bottom']")).Attr,
		html.Attribute{Key: "value", Val: "1"},
	)
	assert.Contains(
		cascadia.Query(poller, selector("input[name='latest_timestamp']")).Attr,
		html.Attribute{Key: "value", Val: "1686025129144"},
	)
}

// Loading the page since a given message
func TestMessagesRoomPollForUpdates(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat detail
	req := httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328?poll&latest_timestamp=1686025129141", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request_with_active_user(req)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".dm-message")), 3)

	// Should have the poller at the bottom
	poller := cascadia.Query(root, selector("#new-messages-poller"))
	assert.NotNil(poller)
	assert.Contains(poller.Attr, html.Attribute{Key: "hx-get", Val: "/messages/1488963321701171204-1178839081222115328"})
	assert.Contains(
		cascadia.Query(poller, selector("input[name='scroll_bottom']")).Attr,
		html.Attribute{Key: "value", Val: "1"},
	)
	assert.Contains(
		cascadia.Query(poller, selector("input[name='latest_timestamp']")).Attr,
		html.Attribute{Key: "value", Val: "1686025129144"},
	)
}

// Loading the page since latest message (no updates)
func TestMessagesRoomPollForUpdatesEmptyResult(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat detail
	req := httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328?poll&latest_timestamp=1686025129144", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request_with_active_user(req)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".dm-message")), 0)

	// Should have the poller at the bottom, with the same value as previously
	poller := cascadia.Query(root, selector("#new-messages-poller"))
	assert.NotNil(poller)
	assert.Contains(poller.Attr, html.Attribute{Key: "hx-get", Val: "/messages/1488963321701171204-1178839081222115328"})
	assert.Contains(
		cascadia.Query(poller, selector("input[name='scroll_bottom']")).Attr,
		html.Attribute{Key: "value", Val: "1"},
	)
	assert.Contains(
		cascadia.Query(poller, selector("input[name='latest_timestamp']")).Attr,
		html.Attribute{Key: "value", Val: "1686025129144"},
	)
}

// Scroll back in the messages
func TestMessagesPaginate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat detail
	req := httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328?cursor=1686025129142", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request_with_active_user(req)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".dm-message")), 2)
}
