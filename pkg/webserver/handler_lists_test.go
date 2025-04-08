package webserver_test

import (
	"fmt"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestListsIndex(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	resp := do_request(httptest.NewRequest("GET", "/lists", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)

	// Check that there's at least 2 Lists
	assert.True(len(cascadia.QueryAll(root, selector(".list-preview"))) >= 2)

	// Check page title
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Lists | Offline Twitter")
}

// Show the users who are on a List
func TestListDetailUsers(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	resp := do_request(httptest.NewRequest("GET", "/lists/1/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 5)

	// Check page title
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Offline Follows | Offline Twitter")
}

// Show the timeline geenrated for a List
func TestListFeed(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	resp := do_request(httptest.NewRequest("GET", "/lists/2", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".timeline > .tweet")), 3)

	// With pagination
	req1 := httptest.NewRequest("GET", "/lists/2?cursor=1629523652000", nil)
	req1.Header.Set("HX-Request", "true")
	resp1 := do_request(req1)
	require.Equal(resp1.StatusCode, 200)
	root2, err := html.Parse(resp1.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root2, selector(":not(.tweet__quoted-tweet) > .tweet")), 2)

	// Check page title
	assert.Equal(cascadia.Query(root, selector("title")).FirstChild.Data, "Bronze Age | Offline Twitter")
}

func TestListFeedInvalidCursor(t *testing.T) {
	require := require.New(t)

	req := httptest.NewRequest("GET", "/lists/2?cursor=asdf", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request(req)
	require.Equal(resp.StatusCode, 400)
}

func TestListDetailDoesntExist(t *testing.T) {
	resp := do_request(httptest.NewRequest("GET", "/lists/2523478", nil))
	require.Equal(t, resp.StatusCode, 404)
}

func TestListDetailInvalidId(t *testing.T) {
	resp := do_request(httptest.NewRequest("GET", "/lists/asd", nil))
	require.Equal(t, resp.StatusCode, 400)
}

func TestListAddAndDeleteUser(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Initial
	resp := do_request(httptest.NewRequest("GET", "/lists/2/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 2)

	// Add a user
	resp_add := do_request(httptest.NewRequest("GET", "/lists/2/add_user?user_handle=cernovich", nil))
	require.Equal(resp_add.StatusCode, 302)
	require.Equal("/lists/2/users", resp_add.Header.Get("Location"))

	// Should be +1 user now
	resp = do_request(httptest.NewRequest("GET", "/lists/2/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 3)

	// Delete a user
	resp_remove := do_request(httptest.NewRequest("GET", "/lists/2/remove_user?user_handle=@cernovich", nil))
	require.Equal(resp_remove.StatusCode, 302)
	require.Equal("/lists/2/users", resp_remove.Header.Get("Location"))

	// Should be +1 user now
	resp = do_request(httptest.NewRequest("GET", "/lists/2/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 2)
}

// Adding invalid users should 400
func TestListAddInvalidUser(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/lists/2/add_user?user_handle=jkwfjekj", nil))
	require.Equal(resp.StatusCode, 400)
}

// Deleting invalid users should 400
func TestListRemoveInvalidUser(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/lists/2/remove_user?user_handle=fwefjkl", nil))
	require.Equal(resp.StatusCode, 400)
}

func TestCreateNewListThenDelete(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Initial list-of-lists
	resp := do_request(httptest.NewRequest("GET", "/lists", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	num_lists := len(cascadia.QueryAll(root, selector(".list-preview")))

	// Create a new list
	resp_add := do_request(httptest.NewRequest("POST", "/lists", strings.NewReader(`{"name": "My New List"}`)))
	require.Equal(resp_add.StatusCode, 302)
	require.Equal(fmt.Sprintf("/lists/%d/users", num_lists+1), resp_add.Header.Get("Location"))

	// Should be N+1 lists now
	resp = do_request(httptest.NewRequest("GET", "/lists", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".list-preview")), num_lists+1)

	// Delete it; should redirect back to Lists index page
	resp_delete := do_request(httptest.NewRequest("DELETE", fmt.Sprintf("/lists/%d", num_lists+1), nil))
	require.Equal(resp_delete.StatusCode, 302)
	require.Equal("/lists", resp_delete.Header.Get("Location"))

	// Should be N lists again
	resp = do_request(httptest.NewRequest("GET", "/lists", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".list-preview")), num_lists)
}
