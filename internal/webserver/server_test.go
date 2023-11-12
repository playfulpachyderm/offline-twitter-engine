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
	app.IsScrapingDisabled = true
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

func TestUserFeedWithEntityInBio(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/michaelmalice", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	bio_entities := cascadia.QueryAll(root, selector(".user-bio .entity"))
	require.Len(bio_entities, 1)
	assert.Equal(bio_entities[0].FirstChild.Data, "@SheathUnderwear")
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

func TestUserFeedTweetsOnlyTab(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/Peter_Nimitz/without_replies", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 2)
}

func TestUserFeedMediaTab(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/Cernovich/media", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 1)
}

func TestUserFeedLikesTab(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/MysteryGrove/likes", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 5)

	// Double check pagination works properly
	resp = do_request(httptest.NewRequest("GET", "/MysteryGrove/likes?cursor=5", nil))
	require.Equal(resp.StatusCode, 200)

	root, err = html.Parse(resp.Body)
	require.NoError(err)
	tweets = cascadia.QueryAll(root, selector(".timeline > .tweet"))
	assert.Len(tweets, 4)
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

func TestSearchWithSortOrder(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/search/think?sort-order=most%20likes", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Contains(cascadia.Query(root, selector("select[name='sort-order'] option[selected]")).FirstChild.Data, "most likes")

	tweets := cascadia.QueryAll(root, selector(".timeline > .tweet"))
	txts := []string{
		"Morally nuanced and complicated discussion",
		"a lot of y’all embarrass yourselves on this",
		"this is why the \"think tank mindset\" is a dead end",
		"At this point what can we expect I guess",
		"Idk if this is relevant to your department",
	}
	for i, txt := range txts {
		assert.Contains(cascadia.Query(tweets[i], selector("p.text")).FirstChild.Data, txt)
	}

	resp = do_request(httptest.NewRequest("GET", "/search/think?sort-order=most%20likes&cursor=413", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	tweets = cascadia.QueryAll(root, selector(".timeline > .tweet"))
	for i, txt := range txts[2:] {
		assert.Contains(cascadia.Query(tweets[i], selector("p.text")).FirstChild.Data, txt)
	}
}

func TestSearchUsers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/search/no?type=users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	user_elements := cascadia.QueryAll(root, selector(".users-list-container .user"))
	assert.Len(user_elements, 2)
}

// Search bar pasted link redirects
// --------------------------------

func TestSearchRedirectOnUserHandle(t *testing.T) {
	assert := assert.New(t)

	resp := do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape("@somebody")), nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/somebody")
}

func TestSearchRedirectOnTweetLink(t *testing.T) {
	assert := assert.New(t)

	// Desktop URL
	resp := do_request(httptest.NewRequest("GET",
		fmt.Sprintf("/search/%s", url.PathEscape("https://twitter.com/wispem_wantex/status/1695221528617468324")),
		nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/tweet/1695221528617468324")

	// Mobile URL
	resp = do_request(httptest.NewRequest("GET",
		fmt.Sprintf("/search/%s", url.PathEscape("https://mobile.twitter.com/wispem_wantex/status/1695221528617468324")),
		nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/tweet/1695221528617468324")
}

func TestSearchRedirectOnUserFeedLink(t *testing.T) {
	assert := assert.New(t)

	// Desktop URL
	resp := do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape("https://twitter.com/agsdf")), nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/agsdf")

	// "With Replies" page
	resp = do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape("https://x.com/agsdf/with_replies")), nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/agsdf")

	// Mobile URL
	resp = do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape("https://mobile.twitter.com/agsdfhh")), nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/agsdfhh")
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

	twt, err := profile.GetTweetById(scraper.TweetID(1695110851324256692))
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

// Lists
// -----

func TestLists(t *testing.T) {
	assert := assert.New(t)
	resp := do_request(httptest.NewRequest("GET", "/lists", nil))
	root, err := html.Parse(resp.Body)
	assert.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list-container .author-info")), 5)
}
