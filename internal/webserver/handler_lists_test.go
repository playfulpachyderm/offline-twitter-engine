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
	require := require.New(t)
	resp := do_request(httptest.NewRequest("GET", "/lists", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)

	// Check that there's at least 2 Lists
	assert.True(t, len(cascadia.QueryAll(root, selector(".list-preview"))) >= 2)
}

func TestListDetail(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Users
	resp := do_request(httptest.NewRequest("GET", "/lists/1/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err := html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 5)

	// Feed
	resp1 := do_request(httptest.NewRequest("GET", "/lists/2", nil))
	require.Equal(resp1.StatusCode, 200)
	root1, err := html.Parse(resp1.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root1, selector(".timeline > .tweet")), 3)
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
	resp_remove := do_request(httptest.NewRequest("GET", "/lists/2/remove_user?user_handle=cernovich", nil))
	require.Equal(resp_remove.StatusCode, 302)
	require.Equal("/lists/2/users", resp_remove.Header.Get("Location"))

	// Should be +1 user now
	resp = do_request(httptest.NewRequest("GET", "/lists/2/users", nil))
	require.Equal(resp.StatusCode, 200)
	root, err = html.Parse(resp.Body)
	require.NoError(err)
	assert.Len(cascadia.QueryAll(root, selector(".users-list .author-info")), 2)
}

func TestCreateNewList(t *testing.T) {
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
}
