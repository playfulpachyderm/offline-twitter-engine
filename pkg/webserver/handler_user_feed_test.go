package webserver_test

import (
	"fmt"
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestUserFeed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/cernovich", nil))
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "@Cernovich | Offline Twitter")

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
	req := httptest.NewRequest("GET", "/cernovich?cursor=1631935701000", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request(req)
	require.Equal(resp.StatusCode, 200)

	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(":not(.tweet__quoted-tweet) > .tweet")), 2)
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

func TestUserFollowersAndFollowees(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	test_cases := []struct {
		Path             string
		NumExpectedUsers int
		ExpectedUsers    []UserID
	}{
		{"/Offline_Twatter/followers", 2, []UserID{1178839081222115328, 1032468021485293568}},
		{"/Offline_Twatter/followees", 1, []UserID{1240784920831762433}},
		{"/Offline_Twatter/mutual_followers", 0, []UserID{}},

		{"/wispem_wantex/followers", 1, []UserID{1240784920831762433}},
		{"/wispem_wantex/followers_you_know", 1, []UserID{1240784920831762433}},
		{"/wispem_wantex/followees", 2, []UserID{1240784920831762433, 358545917}},
		{"/wispem_wantex/followees_you_know", 1, []UserID{1240784920831762433}},
		{"/wispem_wantex/mutual_followers", 1, []UserID{1240784920831762433}},

		{"/cernovich/followers", 1, []UserID{1458284524761075714}},
		{"/cernovich/followers_you_know", 0, []UserID{}},

		{"/schizo_freq/mutual_followers", 1, []UserID{1458284524761075714}},
	}

	for _, test_case := range test_cases {
		resp := do_request_with_active_user(httptest.NewRequest("GET", test_case.Path, nil))
		require.Equal(resp.StatusCode, 200)

		root, err := html.Parse(resp.Body)
		require.NoError(err)
		assert.Len(
			cascadia.QueryAll(root, selector(".users-list > .user")),
			test_case.NumExpectedUsers,
			"Path: %q", test_case.Path,
		)
		for _, u_id := range test_case.ExpectedUsers {
			assert.Len(
				cascadia.QueryAll(root, selector(fmt.Sprintf(".users-list > .user[data-id='%d']", u_id))),
				1,
				"path: %q; id: %d", test_case.Path, u_id,
			)
		}
	}
}
