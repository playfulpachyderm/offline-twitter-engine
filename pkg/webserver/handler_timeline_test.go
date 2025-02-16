package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// TODO: deprecated-offline-follows

func TestTimeline(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/timeline/offline", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Timeline | Offline Twitter")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 21)
}

func TestTimelineWithCursor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	req := httptest.NewRequest("GET", "/timeline/offline?cursor=1631935701000", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request(req)
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweet_nodes := cascadia.QueryAll(root, selector(":not(.tweet__quoted-tweet) > .tweet"))
	assert.Len(tweet_nodes, 10)
}

func TestTimelineWithCursorBadNumber(t *testing.T) {
	require := require.New(t)

	// With a cursor but it sucks
	resp := do_request(httptest.NewRequest("GET", "/timeline/offline?cursor=asdf", nil))
	require.Equal(resp.StatusCode, 400)
}

func TestUserFeedTimeline(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Chat list
	resp := do_request_with_active_user(httptest.NewRequest("GET", "/timeline", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Timeline | Offline Twitter")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 1)
}
