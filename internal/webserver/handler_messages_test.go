package webserver_test

import (
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestMessagesIndexPageRequiresActiveUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// HTMX version
	req := httptest.NewRequest("GET", "/messages", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request(req) // No active user
	require.Equal(401, resp.StatusCode)
	// Piggyback in testing of HTMX 401 toasts
	assert.Equal("beforeend", resp.Header.Get("HX-Reswap"))
	assert.Equal("#toasts", resp.Header.Get("HX-Retarget"))
	assert.Equal("false", resp.Header.Get("HX-Push-Url"))

	// Non-HTMX version
	req1 := httptest.NewRequest("GET", "/messages", nil)
	resp1 := do_request(req1) // No active user
	require.Equal(401, resp1.StatusCode)
	assert.Equal("", resp1.Header.Get("HX-Reswap")) // HX-* stuff should be unset
}

// Loading the index page should work if you're logged in
func TestMessagesIndexPage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat list
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/messages", nil))
	require.Equal(200, resp.StatusCode)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".chat-list .chat-list-entry")), 2)
	assert.Len(cascadia.QueryAll(root, selector(".chat-view .dm-message")), 0) // No messages until you click on one
}

// Users should only be able to open chats they're a member of
func TestMessagesRoomRequiresCorrectUser(t *testing.T) {
	require := require.New(t)

	// No active user
	resp := do_request(httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328", nil))
	require.Equal(401, resp.StatusCode)

	// Wrong user (not in the chat)
	// Copied from `do_request_with_active_user`
	recorder := httptest.NewRecorder()
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.ActiveUser = scraper.User{ID: 782982734, Handle: "Not a real user"} // Simulate a login
	app.WithMiddlewares().ServeHTTP(recorder, httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328", nil))
	resp2 := recorder.Result()
	require.Equal(404, resp2.StatusCode)
}

// Open a chat room
func TestMessagesRoom(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat detail
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328", nil))
	require.Equal(200, resp.StatusCode)
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
	require.Equal(200, resp.StatusCode)
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
	require.Equal(200, resp.StatusCode)
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
	require.Equal(200, resp.StatusCode)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".dm-message")), 2)
}

// When scraping is disabled, marking as read should 401
func TestMessagesMarkAsRead(t *testing.T) {
	require := require.New(t)

	resp := do_request_with_active_user(httptest.NewRequest("GET", "/messages/1488963321701171204-1178839081222115328/mark-as-read", nil))
	require.Equal(resp.StatusCode, 401)
}

// When scraping is disabled, sending a message should 401
func TestMessagesSend(t *testing.T) {
	require := require.New(t)

	resp := do_request_with_active_user(httptest.NewRequest("GET",
		"/messages/1488963321701171204-1178839081222115328/send",
		strings.NewReader(`{"text": "bleh"}`),
	))
	require.Equal(401, resp.StatusCode)
}

// When scraping is disabled, sending a reacc should 401
func TestMessagesSendReacc(t *testing.T) {
	require := require.New(t)

	resp := do_request_with_active_user(httptest.NewRequest("GET",
		"/messages/1488963321701171204-1178839081222115328/reacc",
		strings.NewReader(`{"message_id": "1", "reacc": ":)"}`),
	))
	require.Equal(401, resp.StatusCode)
}

func TestMessagesRefreshConversationsList(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// No active chat
	req := httptest.NewRequest("GET", "/messages/refresh-list", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request_with_active_user(req)
	require.Equal(200, resp.StatusCode)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".chat-list-entry")), 2)
	assert.Len(cascadia.QueryAll(root, selector(".chat-list-entry.chat-list-entry--active-chat")), 0)

	// With an active chat
	req1 := httptest.NewRequest("GET", "/messages/refresh-list?active-chat=1488963321701171204-1178839081222115328", nil)
	req1.Header.Set("HX-Request", "true")
	resp1 := do_request_with_active_user(req1)
	require.Equal(200, resp1.StatusCode)
	root1, err := html.Parse(resp1.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root1, selector(".chat-list-entry")), 2)
	assert.Len(cascadia.QueryAll(root1, selector(".chat-list-entry.chat-list-entry--active-chat")), 1)
}
