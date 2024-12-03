package webserver_test

import (
	"fmt"
	"net/url"
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestSearchQueryStringRedirect(t *testing.T) {
	assert := assert.New(t)

	resp := do_request(httptest.NewRequest("GET", "/search?q=asdf", nil))
	assert.Equal(resp.StatusCode, 302)
	assert.Equal(resp.Header.Get("Location"), "/search/asdf")
}

func TestSearch(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	search_txt := "to:spacex to:covfefeanon"

	resp := do_request(httptest.NewRequest("GET", fmt.Sprintf("/search/%s", url.PathEscape(search_txt)), nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "Search | Offline Twitter")
	assert.Contains(cascadia.Query(root, selector("#searchBar")).Attr, html.Attribute{Key: "value", Val: search_txt})

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
	resp = do_request(httptest.NewRequest("GET", "/search/who%20are?cursor=1628979529000", nil))
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
	user_elements := cascadia.QueryAll(root, selector(".users-list .user"))
	assert.Len(user_elements, 2)
	assert.Contains(cascadia.Query(root, selector("#searchBar")).Attr, html.Attribute{Key: "value", Val: "no"})
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
