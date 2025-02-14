package persistence_test

import (
	"testing"

	"fmt"
	"math/rand"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestSaveAndLoadOfflineList(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create an offline list
	l := List{IsOnline: false, Name: fmt.Sprintf("Test List %d", rand.Int())}
	require.Equal(l.ID, ListID(0))
	profile.SaveList(&l)
	require.NotEqual(l.ID, ListID(0)) // ID should be assigned when it's saved

	// Check it comes back the same
	new_l, err := profile.GetListById(l.ID)
	require.NoError(err)
	assert.Equal(l.ID, new_l.ID)
	assert.Equal(l.IsOnline, new_l.IsOnline)
	assert.Equal(l.Name, new_l.Name)
}

func TestRenameOfflineList(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create an offline list
	l := List{IsOnline: false, Name: fmt.Sprintf("Test List %d", rand.Int())}
	profile.SaveList(&l)
	require.NotEqual(l.ID, ListID(0))

	// Rename it
	l.Name = fmt.Sprintf("Untest List %d", rand.Int())
	profile.SaveList(&l)

	// Rename should be effective
	new_l, err := profile.GetListById(l.ID)
	require.NoError(err)
	assert.Equal(l.ID, new_l.ID)
	assert.Equal(l.IsOnline, new_l.IsOnline)
	assert.Equal(l.Name, new_l.Name)
}

func TestSaveAndLoadOnlineList(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create an online list
	l := List{IsOnline: true, OnlineID: OnlineListID(rand.Int()), Name: fmt.Sprintf("Test List %d", rand.Int())}
	require.Equal(l.ID, ListID(0))
	profile.SaveList(&l)
	require.NotEqual(l.ID, ListID(0)) // ID should be assigned when it's saved

	// Check it comes back the same
	new_l, err := profile.GetListById(l.ID)
	require.NoError(err)
	assert.Equal(l.ID, new_l.ID)
	assert.Equal(l.IsOnline, new_l.IsOnline)
	assert.Equal(l.OnlineID, new_l.OnlineID) // Check OnlineID for online lists
	assert.Equal(l.Name, new_l.Name)
}

func TestRenameOnlineList(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create an online list
	l := List{IsOnline: true, OnlineID: OnlineListID(rand.Int()), Name: fmt.Sprintf("Test List %d", rand.Int())}
	profile.SaveList(&l)
	require.NotEqual(l.ID, ListID(0))

	// Rename it
	l.Name = fmt.Sprintf("Untest List %d", rand.Int())
	profile.SaveList(&l)

	// Rename should be effective
	new_l, err := profile.GetListById(l.ID)
	require.NoError(err)
	assert.Equal(l.ID, new_l.ID)
	assert.Equal(l.IsOnline, new_l.IsOnline)
	assert.Equal(l.OnlineID, new_l.OnlineID) // Check OnlineID for online lists
	assert.Equal(l.Name, new_l.Name)
}

func TestNoOnlineListWithoutID(t *testing.T) {
	require := require.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Creating an online list with no OnlineID should fail
	l := List{IsOnline: true, OnlineID: OnlineListID(0), Name: fmt.Sprintf("Test List %d", rand.Int())}
	defer func() {
		// Assert a panic occurred
		r, is_ok := recover().(error)
		require.True(is_ok)
		require.Error(r)
	}()
	profile.SaveList(&l)
}

func TestAddAndRemoveUserToList(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create a list
	l := List{IsOnline: false, Name: fmt.Sprintf("Test List %d", rand.Int())}
	profile.SaveList(&l)

	// Check there's no users in it
	require.Len(profile.GetListUsers(l.ID), 0)

	// Add a user to the list
	u := create_dummy_user()
	require.NoError(profile.SaveUser(&u))
	profile.SaveListUser(l.ID, u.ID)

	// Make sure it's in the list
	users := profile.GetListUsers(l.ID)
	require.Len(users, 1)
	assert.Equal(u.Handle, users[0].Handle)

	// Addding it again should do nothing
	profile.SaveListUser(l.ID, u.ID)
	require.Len(profile.GetListUsers(l.ID), 1)

	// Remove the user from the list
	profile.DeleteListUser(l.ID, u.ID)

	// Should be gone
	require.Len(profile.GetListUsers(l.ID), 0)
}

func TestGetAllLists(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create a list
	l := List{IsOnline: false, Name: fmt.Sprintf("Test List %d", rand.Int())}
	profile.SaveList(&l)

	// Get all the lists
	lists := profile.GetAllLists()
	assert.True(len(lists) > 1) // Should be at least Offline Follows and `l`
	assert.Contains(lists, l)
}

func TestDeleteList(t *testing.T) {
	require := require.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Create an offline list
	l := List{IsOnline: false, Name: fmt.Sprintf("Test List %d", rand.Int())}
	require.Equal(l.ID, ListID(0))
	profile.SaveList(&l)
	require.NotEqual(l.ID, ListID(0)) // ID should be assigned when it's saved

	// Delete it
	profile.DeleteList(l.ID)

	// Check it's gone
	_, err = profile.GetListById(l.ID)
	require.Error(err)
}
