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
	assert.Len(tweet_nodes, 20)
}

func TestTimelineWithCursor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/timeline/offline?cursor=1631935701000", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Timeline | Offline Twitter")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
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

	// Boilerplate for setting an active user
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.ActiveUser = scraper.User{ID: 1488963321701171204, Handle: "Offline_Twatter"} // Simulate a login

	// Chat list
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest("GET", "/timeline", nil))
	resp := recorder.Result()
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Timeline | Offline Twitter")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 1)
}
