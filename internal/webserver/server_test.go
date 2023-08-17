package webserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type CapturingWriter struct {
	Writes [][]byte
}

func (w *CapturingWriter) Write(p []byte) (int, error) {
	w.Writes = append(w.Writes, p)
	return len(p), nil
}

var profile persistence.Profile

func init() {
	var err error
	profile, err = persistence.LoadProfile("../../sample_data/profile")
	if err != nil {
		panic(err)
	}
}

func selector(s string) cascadia.Sel {
	ret, err := cascadia.Parse(s)
	if err != nil {
		panic(err)
	}
	return ret
}

func do_request(req *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	app := webserver.NewApp(profile)
	app.DisableScraping = true
	app.ServeHTTP(recorder, req)
	return recorder.Result()
}

// Homepage
// --------

func TestHomepage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | Home")
}

// User feed
// ---------

func TestUserFeed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/cernovich", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | @Cernovich")

	tweet_nodes := cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(tweet_nodes, 7)
}

func TestUserFeedMissing(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/awefhwefhwejh", nil))
	require.Equal(resp.StatusCode, 404)
}

func TestUserFeedWithCursor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// With a cursor
	resp := do_request(httptest.NewRequest("GET", "/cernovich?cursor=1631935701", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | @Cernovich")

	tweet_nodes := cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(tweet_nodes, 2)
}

func TestUserFeedWithCursorBadNumber(t *testing.T) {
	require := require.New(t)

	// With a cursor but it sucks
	resp := do_request(httptest.NewRequest("GET", "/cernovich?cursor=asdf", nil))
	require.Equal(resp.StatusCode, 400)
}

// Tweet Detail page
// -----------------

func TestTweetDetail(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1413773185296650241", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweet_nodes := cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(tweet_nodes, 4)
}

func TestTweetDetailMissing(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/100089", nil))
	require.Equal(resp.StatusCode, 404)
}

func TestTweetDetailInvalidNumber(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/fwjgkj", nil))
	require.Equal(resp.StatusCode, 400)
}

// Static content
// --------------

func TestStaticFile(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/static/styles.css", nil))
	require.Equal(resp.StatusCode, 200)
}

func TestStaticFileNonexistent(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/static/blehblehblehwfe", nil))
	require.Equal(resp.StatusCode, 404)
}
