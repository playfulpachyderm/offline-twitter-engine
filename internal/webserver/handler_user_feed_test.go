package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestUserFeed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/cernovich", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "@Cernovich | Offline Twitter")

	assert.Len(cascadia.QueryAll(root, selector(".timeline > .tweet")), 8)
	assert.Len(cascadia.QueryAll(root, selector(".timeline > .pinned-tweet")), 1)
	assert.Len(cascadia.QueryAll(root, selector(".tweet")), 12) // Pinned tweet appears again
}

func TestUserFeedWithEntityInBio(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/michaelmalice", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	bio_entities := cascadia.QueryAll(root, selector(".user-header__bio .entity"))
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
	resp := do_request(httptest.NewRequest("GET", "/cernovich?cursor=1631935701000", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	title_node := cascadia.Query(root, selector("title"))
	assert.Equal(title_node.FirstChild.Data, "@Cernovich | Offline Twitter")

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

// Followers and followees
// -----------------------

func TestUserFollowers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/Offline_Twatter/followers", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list > .user")), 2)
}

func TestUserFollowees(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/Offline_Twatter/followees", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list > .user")), 1)
}
