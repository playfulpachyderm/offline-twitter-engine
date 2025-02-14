package persistence_test

import (
	"testing"

	"fmt"
	"math/rand"
	"os"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// Create a user, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadUser(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	fake_user := create_dummy_user()

	// Save the user, then reload it and ensure it's the same
	err := profile.SaveUser(&fake_user)
	require.NoError(err)

	new_fake_user, err := profile.GetUserByID(fake_user.ID)
	require.NoError(err)

	if diff := deep.Equal(new_fake_user, fake_user); diff != nil {
		t.Error(diff)
	}

	// Same thing, but get by handle
	new_fake_user2, err := profile.GetUserByHandle(fake_user.Handle)
	require.NoError(err)

	if diff := deep.Equal(new_fake_user2, fake_user); diff != nil {
		t.Error(diff)
	}
}

func TestModifyUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()
	user.DisplayName = "Display Name 1"
	user.Location = "location1"
	user.Handle = UserHandle(fmt.Sprintf("handle %d", rand.Int31()))
	user.IsPrivate = false
	user.IsVerified = false
	user.FollowersCount = 1000
	user.JoinDate = TimestampFromUnix(1000)
	user.ProfileImageUrl = "asdf"
	user.IsContentDownloaded = true

	// Save the user for the first time; should do insert
	err := profile.SaveUser(&user)
	require.NoError(err)

	new_handle := UserHandle(fmt.Sprintf("handle %d", rand.Int31()))

	user.DisplayName = "Display Name 2"
	user.Location = "location2"
	user.Handle = new_handle
	user.IsPrivate = true
	user.IsVerified = true
	user.FollowersCount = 2000
	user.JoinDate = TimestampFromUnix(2000)
	user.ProfileImageUrl = "asdf2"
	user.IsContentDownloaded = false // test No Worsening

	// Save the user for the second time; should do update
	err = profile.SaveUser(&user)
	require.NoError(err)

	// Reload the modified user
	new_user, err := profile.GetUserByID(user.ID)
	require.NoError(err)

	assert.Equal("Display Name 2", new_user.DisplayName)
	assert.Equal(new_handle, new_user.Handle)
	assert.Equal("location2", new_user.Location)
	assert.True(new_user.IsPrivate)
	assert.True(new_user.IsVerified)
	assert.Equal(2000, new_user.FollowersCount)
	assert.Equal(int64(1000), new_user.JoinDate.Unix())
	assert.Equal("asdf2", new_user.ProfileImageUrl)
	assert.True(new_user.IsContentDownloaded)
}

func TestSetUserBannedDeleted(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()
	user.DisplayName = "Display Name 1"
	user.Location = "location1"
	user.Bio = "Some Bio"
	user.Handle = UserHandle(fmt.Sprintf("handle %d", rand.Int31()))
	user.FollowersCount = 1000
	user.JoinDate = TimestampFromUnix(1000)
	user.ProfileImageUrl = "asdf"
	user.IsContentDownloaded = true

	// Save the user so it can be modified
	err := profile.SaveUser(&user)
	require.NoError(err)

	// Now the user deactivates
	err = profile.SaveUser(&User{ID: user.ID, IsDeleted: true})
	require.NoError(err)
	// Reload the modified user
	new_user, err := profile.GetUserByID(user.ID)
	require.NoError(err)

	assert.True(new_user.IsDeleted)
	assert.Equal(new_user.Handle, user.Handle)
	assert.Equal(new_user.DisplayName, user.DisplayName)
	assert.Equal(new_user.Location, user.Location)
	assert.Equal(new_user.Bio, user.Bio)
	assert.Equal(new_user.FollowersCount, user.FollowersCount)
	assert.Equal(new_user.JoinDate, user.JoinDate)
	assert.Equal(new_user.ProfileImageUrl, user.ProfileImageUrl)
	assert.Equal(new_user.IsContentDownloaded, user.IsContentDownloaded)
}

func TestSaveAndLoadBannedDeletedUser(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := User{
		ID:       UserID(rand.Int31()),
		Handle:   UserHandle(fmt.Sprintf("handle-%d", rand.Int31())),
		IsBanned: true,
	}

	// Save the user, then reload it and ensure it's the same
	err := profile.SaveUser(&user)
	require.NoError(err)
	new_user, err := profile.GetUserByID(user.ID)
	require.NoError(err)

	if diff := deep.Equal(new_user, user); diff != nil {
		t.Error(diff)
	}
}

func TestInsertUserWithDuplicateHandle(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user1 := create_dummy_user()
	err := profile.SaveUser(&user1)
	require.NoError(err)

	user2 := create_dummy_user()
	user2.Handle = user1.Handle
	err = profile.SaveUser(&user2)
	var conflict_err ErrConflictingUserHandle
	require.ErrorAs(err, &conflict_err)
	require.Equal(conflict_err.ConflictingUserID, user1.ID)

	// Should have inserted user2
	new_user2, err := profile.GetUserByID(user2.ID)
	require.NoError(err)
	if diff := deep.Equal(user2, new_user2); diff != nil {
		t.Error(diff)
	}

	// user1 should be marked as deleted
	new_user1, err := profile.GetUserByID(user1.ID)
	require.NoError(err)
	user1.IsDeleted = true // for equality check
	if diff := deep.Equal(user1, new_user1); diff != nil {
		t.Error(diff)
	}
}

func TestReviveDeletedUserWithDuplicateHandle(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	// Create the first user (deleted)
	user1 := create_dummy_user()
	user1.DisplayName = "user1"
	user1.IsDeleted = true
	err := profile.SaveUser(&user1)
	require.NoError(err)

	// Create the second user (active)
	user2 := create_dummy_user()
	user2.Handle = user1.Handle
	user2.DisplayName = "user2"
	err = profile.SaveUser(&user2)
	require.NoError(err)

	// Reactivate the 1st user; should return the 2nd user's ID as a conflict
	user1.IsDeleted = false
	err = profile.SaveUser(&user1)
	var conflict_err ErrConflictingUserHandle
	require.ErrorAs(err, &conflict_err)
	require.Equal(conflict_err.ConflictingUserID, user2.ID)

	// User1 should be updated (no longer deleted)
	new_user1, err := profile.GetUserByID(user1.ID)
	require.NoError(err)
	assert.False(new_user1.IsDeleted)
	// User2 should be marked deleted
	new_user2, err := profile.GetUserByID(user2.ID)
	require.NoError(err)
	assert.True(new_user2.IsDeleted)
}

// `profile.GetUserByHandle` should be case-insensitive
func TestHandleIsCaseInsensitive(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_stable_user()

	new_user, err := profile.GetUserByHandle("hANdle StaBlE")
	require.NoError(t, err, "Couldn't find the user")

	if diff := deep.Equal(user, new_user); diff != nil {
		t.Error(diff)
	}
}

// Test scenarios relating to user content downloading
func TestCheckUserContentDownloadNeeded(t *testing.T) {
	assert := assert.New(t)
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()

	// If user is not in database, should be "yes" automatically
	assert.True(profile.CheckUserContentDownloadNeeded(user))

	// Save the user, but `is_content_downloaded` is still "false"
	user.BannerImageUrl = "banner url1"
	user.ProfileImageUrl = "profile url1"
	user.IsContentDownloaded = false
	err := profile.SaveUser(&user)
	require.NoError(t, err)

	// If is_content_downloaded is false, then it needs download
	assert.True(profile.CheckUserContentDownloadNeeded(user))

	// Mark `is_content_downloaded` as "true" again
	user.IsContentDownloaded = true
	err = profile.SaveUser(&user)
	require.NoError(t, err)

	// If the file exists, then it should not require any download
	_, err = os.Create("test_profiles/TestUserQueries/profile_images/" + user.BannerImageLocalPath)
	require.NoError(t, err)
	_, err = os.Create("test_profiles/TestUserQueries/profile_images/" + user.ProfileImageLocalPath)
	require.NoError(t, err)
	assert.False(profile.CheckUserContentDownloadNeeded(user))

	// Download needed for new banner image
	user.BannerImageLocalPath = "banner url2"
	assert.True(profile.CheckUserContentDownloadNeeded(user))
}

// Make sure following works
//
// - users are unfollowed by default
// - following a user makes it save as is_followed
// - using regular save method doesn't un-follow
// - unfollowing a user makes it save as no longer is_followed
func TestFollowUnfollowUser(t *testing.T) {
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()
	assert.False(user.IsFollowed)
	err := profile.SaveUser(&user)
	assert.NoError(err)

	profile.SetUserFollowed(&user, true)
	assert.True(user.IsFollowed)

	// Ensure the change was persisted
	user_reloaded, err := profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID) // Verify it's the same user
	assert.True(user_reloaded.IsFollowed)

	err = profile.SaveUser(&user) // should NOT un-set is_followed
	assert.NoError(err)
	user_reloaded, err = profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID) // Verify it's the same user
	assert.True(user_reloaded.IsFollowed)

	profile.SetUserFollowed(&user, false)
	assert.False(user.IsFollowed)

	// Ensure the change was persisted
	user_reloaded, err = profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID) // Verify it's the same user
	assert.False(user_reloaded.IsFollowed)
}

// Should correctly report whether a User is followed or not, according to the DB (not the in-memory objects)
func TestIsFollowingUser(t *testing.T) {
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	// Create the user
	user := create_dummy_user()
	assert.False(user.IsFollowed)
	assert.False(profile.IsFollowing(user))

	err := profile.SaveUser(&user)
	assert.NoError(err)

	// Make sure the user isn't "followed"
	assert.False(profile.IsFollowing(user))
	user.IsFollowed = true
	assert.False(profile.IsFollowing(user)) // Should check the DB not the in-memory User
	user.IsFollowed = false

	profile.SetUserFollowed(&user, true)

	assert.True(profile.IsFollowing(user))
	user.IsFollowed = false
	assert.True(profile.IsFollowing(user)) // Check the DB, not the User
	user.IsFollowed = true

	profile.SetUserFollowed(&user, false)
	assert.False(profile.IsFollowing(user))
}

// Should create a new Unknown User from the given handle.
// The Unknown User should work consistently with other Users.
func TestCreateUnknownUserWithHandle(t *testing.T) {
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	next_id := profile.NextFakeUserID()

	handle := UserHandle(fmt.Sprintf("UnknownUser%d", rand.Int31()))
	user := GetUnknownUserWithHandle(handle)
	assert.Equal(UserID(0), user.ID)
	assert.True(user.IsIdFake)

	err := profile.SaveUser(&user)
	assert.NoError(err)
	assert.Equal(UserID(next_id+1), user.ID)

	// Ensure the change was persisted
	user_reloaded, err := profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(handle, user_reloaded.Handle) // Verify it's the same user
	assert.Equal(UserID(next_id+1), user_reloaded.ID)

	// Why not tack this test on here: make sure NextFakeUserID works as expected
	assert.Equal(next_id+2, profile.NextFakeUserID())
}

// Should update the unknown User's UserID with the correct ID if it already exists
func TestCreateUnknownUserWithHandleThatAlreadyExists(t *testing.T) {
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_stable_user()

	unknown_user := GetUnknownUserWithHandle(user.Handle)
	assert.Equal(UserID(0), unknown_user.ID)

	err := profile.SaveUser(&unknown_user)
	assert.NoError(err)
	assert.Equal(user.ID, unknown_user.ID)

	// The real user should not have been overwritten at all
	user_reloaded, err := profile.GetUserByID(user.ID)
	assert.NoError(err)
	assert.False(user_reloaded.IsIdFake) // This one particularly
	assert.Equal(user.Handle, user_reloaded.Handle)
	assert.Equal(user.Bio, user_reloaded.Bio)
	assert.Equal(user.DisplayName, user_reloaded.DisplayName)
}

func TestSearchUsers(t *testing.T) {
	assert := assert.New(t)

	profile_path := "../../sample_data/profile"
	profile := create_or_load_profile(profile_path)

	users := profile.SearchUsers("no")
	assert.Len(users, 2)
	assert.Equal(users[0].Handle, UserHandle("Cernovich"))
	assert.Equal(users[1].Handle, UserHandle("CovfefeAnon"))
}
