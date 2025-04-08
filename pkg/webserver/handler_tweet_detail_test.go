package webserver_test

import (
	"io"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestTweetDetail(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1413773185296650241", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweet_nodes := cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(tweet_nodes, 4)
	// Check page title
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Tweet | Offline Twitter")
}

func TestTweetDetailMissing(t *testing.T) {
	require := require.New(t)

	// Suppress error logging in console
	recorder := httptest.NewRecorder()
	app := make_testing_app(nil)
	app.ErrorLog.SetOutput(io.Discard)
	app.WithMiddlewares().ServeHTTP(recorder, httptest.NewRequest("GET", "/tweet/100089", nil))
	resp := recorder.Result()
	require.Equal(resp.StatusCode, 404)
}

func TestTweetDetailInvalidNumber(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/fwjgkj", nil))
	require.Equal(resp.StatusCode, 400)
}

func TestTweetsWithContent(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Poll
	resp := do_request(httptest.NewRequest("GET", "/tweet/1465534109573390348", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".poll")), 1)
	assert.Len(cascadia.QueryAll(root, selector(".poll__choice")), 4)

	// Video
	resp = do_request(httptest.NewRequest("GET", "/tweet/1453461248142495744", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector("video")), 1)

	// Url
	resp = do_request(httptest.NewRequest("GET", "/tweet/1438642143170646017", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".embedded-link")), 3)

	// Space
	resp = do_request(httptest.NewRequest("GET", "/tweet/1624833173514293249", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".space")), 1)
	assert.Len(cascadia.QueryAll(root, selector("ul.space__participants-list li")), 9)
}

func TestTweetWithEntities(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1489944024278523906", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	entities := cascadia.QueryAll(root, selector(".entity"))
	assert.Len(entities, 2)
	assert.Equal(entities[0].Data, "a")
	assert.Equal(entities[0].FirstChild.Data, "@gofundme")
	assert.Contains(entities[0].Attr, html.Attribute{Key: "href", Val: "/gofundme"})
	assert.Equal(entities[1].Data, "a")
	assert.Equal(entities[1].FirstChild.Data, "#BankruptGoFundMe")
	assert.Contains(entities[1].Attr, html.Attribute{Key: "href", Val: "/search/%23BankruptGoFundMe"})
}

func TestLongTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1695110851324256692", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	paragraphs := cascadia.QueryAll(root, selector(".tweet .text"))
	assert.Len(paragraphs, 22)

	twt, err := profile.GetTweetById(TweetID(1695110851324256692))
	require.NoError(err)
	for i, s := range strings.Split(twt.Text, "\n") {
		assert.Equal(strings.TrimSpace(s), strings.TrimSpace(paragraphs[i].FirstChild.Data))
	}
}

func TestTombstoneTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/31", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tombstone := cascadia.Query(root, selector(".tweet .tombstone"))
	assert.Equal("This Tweet was deleted by the Tweet author", strings.TrimSpace(tombstone.FirstChild.Data))
}

func TestTweetThread(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1698762403163304110", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)

	reply_chains := cascadia.QueryAll(root, selector(".reply-chain"))
	require.Len(reply_chains, 2)

	thread_chain := reply_chains[0]
	assert.Len(cascadia.QueryAll(thread_chain, selector(".reply-tweet")), 7)
}
