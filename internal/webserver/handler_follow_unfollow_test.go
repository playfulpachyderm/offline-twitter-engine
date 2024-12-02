package webserver_test

import (
	"testing"

	"net/http/httptest"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// TODO: deprecated-offline-follows

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
