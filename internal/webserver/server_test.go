package webserver_test

import (
	"testing"

	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
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

// Should redirect to the timeline
func TestHomepage(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/", nil))
	require.Equal(resp.StatusCode, 303)
	require.Equal(resp.Header.Get("Location"), "/timeline")
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

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 7)
	including_quote_tweets := cascadia.QueryAll(root, selector(".tweet"))
	assert.Len(including_quote_tweets, 10)
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

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 2)
}

func TestUserFeedWithCursorBadNumber(t *testing.T) {
	require := require.New(t)

	// With a cursor but it sucks
	resp := do_request(httptest.NewRequest("GET", "/cernovich?cursor=asdf", nil))
	require.Equal(resp.StatusCode, 400)
}

// Timeline page
// -------------

func TestTimeline(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/timeline", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | Timeline")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 18)
}

func TestTimelineWithCursor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/timeline?cursor=1631935701", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | Timeline")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 10)
}

func TestTimelineWithCursorBadNumber(t *testing.T) {
	require := require.New(t)

	// With a cursor but it sucks
	resp := do_request(httptest.NewRequest("GET", "/timeline?cursor=asdf", nil))
	require.Equal(resp.StatusCode, 400)
}

// Search page
// -----------

func TestSearchQueryStringRedirect(t *testing.T) {
	assert := assert.New(t)

	// With a cursor but it sucks
	resp := do_request(httptest.NewRequest("GET", "/search?q=asdf", nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/search/asdf")
}

func TestSearch(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape("to:spacex to:covfefeanon")), nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Offline Twitter | Search")

	tweet_nodes := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweet_nodes, 1)
}

func TestSearchWithCursor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// First, without the cursor
	resp := do_request(httptest.NewRequest("GET", "/search/who%20are", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".timeline > .tweet")), 3)

	// Add a cursor with the 1st tweet's posted_at time
	resp = do_request(httptest.NewRequest("GET", "/search/who%20are?cursor=1628979529", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".timeline > .tweet")), 2)
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

func TestTweetsWithContent(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Poll
	resp := do_request(httptest.NewRequest("GET", "/tweet/1465534109573390348", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".poll")), 1)
	assert.Len(cascadia.QueryAll(root, selector(".poll-choice")), 4)

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
	assert.Len(cascadia.QueryAll(root, selector("ul.space-participants-list li")), 9)
}

func TestLongTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/tweet/1695110851324256692", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	paragraphs := cascadia.QueryAll(root, selector(".tweet-text"))
	assert.Len(paragraphs, 22)

	twt, err := profile.GetTweetById(scraper.TweetID(1695110851324256692))
	require.NoError(err)
	for i, s := range strings.Split(twt.Text, "\n") {
		assert.Equal(strings.TrimSpace(s), strings.TrimSpace(paragraphs[i].FirstChild.Data))
	}
}

// Follow and unfollow
// -------------------

func TestFollowUnfollow(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	user, err := profile.GetUserByHandle("kwamurai")
	require.NoError(err)
	require.False(user.IsFollowed)

	// Follow the user
	resp := do_request(httptest.NewRequest("POST", "/follow/kwamurai", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	button := cascadia.Query(root, selector("button"))
	assert.Contains(button.Attr, html.Attribute{Key: "hx-post", Val: "/unfollow/kwamurai"})
	assert.Equal(strings.TrimSpace(button.FirstChild.Data), "Unfollow")

	user, err = profile.GetUserByHandle("kwamurai")
	require.NoError(err)
	require.True(user.IsFollowed)

	// Unfollow the user
	resp = do_request(httptest.NewRequest("POST", "/unfollow/kwamurai", nil))
	require.Equal(resp.StatusCode, 200)

	root, err = html.Parse(resp.Body)
	require.NoError(err)
	button = cascadia.Query(root, selector("button"))
	assert.Contains(button.Attr, html.Attribute{Key: "hx-post", Val: "/follow/kwamurai"})
	assert.Equal(strings.TrimSpace(button.FirstChild.Data), "Follow")

	user, err = profile.GetUserByHandle("kwamurai")
	require.NoError(err)
	require.False(user.IsFollowed)
}

func TestFollowUnfollowPostOnly(t *testing.T) {
	require := require.New(t)
	resp := do_request(httptest.NewRequest("GET", "/follow/kwamurai", nil))
	require.Equal(resp.StatusCode, 405)
	resp = do_request(httptest.NewRequest("GET", "/unfollow/kwamurai", nil))
	require.Equal(resp.StatusCode, 405)
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
