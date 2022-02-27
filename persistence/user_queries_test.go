package persistence_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"
)


/**
 * Create a user, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadUser(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	fake_user := create_dummy_user()

	// Save the user, then reload it and ensure it's the same
	err := profile.SaveUser(fake_user)
	if err != nil {
		panic(err)
	}
	new_fake_user, err := profile.GetUserByID(fake_user.ID)
	if err != nil {
		panic(err)
	}

	if diff := deep.Equal(new_fake_user, fake_user); diff != nil {
		t.Error(diff)
	}

	// Same thing, but get by handle
	new_fake_user2, err := profile.GetUserByHandle(fake_user.Handle)
	if err != nil {
		panic(err)
	}

	if diff := deep.Equal(new_fake_user2, fake_user); diff != nil {
		t.Error(diff)
	}
}

/**
 *
 */
func TestModifyUser(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	fake_user := create_dummy_user()
	fake_user.DisplayName = "Display Name 1"
	fake_user.Location = "location1"
	fake_user.IsPrivate = false
	fake_user.IsVerified = false
	fake_user.IsBanned = false
	fake_user.FollowersCount = 1000
	fake_user.JoinDate = time.Unix(1000, 0)
	fake_user.ProfileImageUrl = "asdf"
	fake_user.IsContentDownloaded = true

	// Save the user so it can be modified
	err := profile.SaveUser(fake_user)
	if err != nil {
		panic(err)
	}


	fake_user.DisplayName = "Display Name 2"
	fake_user.Location = "location2"
	fake_user.IsPrivate = true
	fake_user.IsVerified = true
	fake_user.IsBanned = true
	fake_user.FollowersCount = 2000
	fake_user.JoinDate = time.Unix(2000, 0)
	fake_user.ProfileImageUrl = "asdf2"
	fake_user.IsContentDownloaded = false  // test No Worsening

	// Save the modified user
	err = profile.SaveUser(fake_user)
	if err != nil {
		panic(err)
	}
	// Reload the modified user
	new_fake_user, err := profile.GetUserByID(fake_user.ID)
	if err != nil {
		panic(err)
	}

	if new_fake_user.DisplayName != "Display Name 2" {
		t.Errorf("Expected display name %q, got %q", "Display Name 2", new_fake_user.DisplayName)
	}
	if new_fake_user.Location != "location2" {
		t.Errorf("Expected location %q, got %q", "location2", new_fake_user.Location)
	}
	if new_fake_user.IsPrivate != true {
		t.Errorf("Should be private")
	}
	if new_fake_user.IsVerified != true {
		t.Errorf("Should be verified")
	}
	if new_fake_user.IsBanned != true {
		t.Errorf("Should be banned")
	}
	if new_fake_user.FollowersCount != 2000 {
		t.Errorf("Expected %d followers, got %d", 2000, new_fake_user.FollowersCount)
	}
	if new_fake_user.JoinDate.Unix() != 1000 {
		t.Errorf("Expected unchanged join date (%d), got %d", 1000, new_fake_user.JoinDate.Unix())
	}
	if new_fake_user.ProfileImageUrl != "asdf2" {
		t.Errorf("Expected profile image url to be %q, got %q", "asdf2", new_fake_user.ProfileImageUrl)
	}
	if new_fake_user.IsContentDownloaded != true {
		t.Errorf("Expected content to be downloaded (no-worsening)")
	}
}

func TestHandleIsCaseInsensitive(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_stable_user()

	new_user, err := profile.GetUserByHandle("hANdle StaBlE")
	if err != nil {
		t.Fatalf("Couldn't find the user: %s", err.Error())
	}

	if diff := deep.Equal(user, new_user); diff != nil {
		t.Error(diff)
	}
}


/**
 * Should correctly report whether the user exists in the database
 */
func TestUserExists(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()

	exists := profile.UserExists(user.Handle)
	if exists {
		t.Errorf("It shouldn't exist, but it does: %d", user.ID)
	}
	err := profile.SaveUser(user)
	if err != nil {
		panic(err)
	}
	exists = profile.UserExists(user.Handle)
	if !exists {
		t.Errorf("It should exist, but it doesn't: %d", user.ID)
	}
}

/**
 * Test scenarios relating to user content downloading
 */
func TestCheckUserContentDownloadNeeded(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()

	// If user is not in database, should be "yes" automatically
	if profile.CheckUserContentDownloadNeeded(user) != true {
		t.Errorf("Non-saved user should always require content download")
	}

	// Save the user, but `is_content_downloaded` is still "false"
	user.BannerImageUrl = "banner url1"
	user.ProfileImageUrl = "profile url1"
	user.IsContentDownloaded = false
	err := profile.SaveUser(user)
	if err != nil {
		panic(err)
	}

	// If is_content_downloaded is false, then it needs download
	if profile.CheckUserContentDownloadNeeded(user) != true {
		t.Errorf("Non-downloaded user should require download")
	}

	// Mark `is_content_downloaded` as "true" again
	user.IsContentDownloaded = true
	err = profile.SaveUser(user)
	if err != nil {
		panic(err)
	}

	// If everything is up to date, no download should be required
	if profile.CheckUserContentDownloadNeeded(user) != false {
		t.Errorf("Up-to-date user shouldn't need a download")
	}

	// Change an URL, but don't save it-- needs to be different from what's in the DB
	user.BannerImageUrl = "banner url2"

	// Download needed for new banner image
	if profile.CheckUserContentDownloadNeeded(user) != true {
		t.Errorf("If banner image changed, user should require another download")
	}
}

/**
 * Make sure following works
 *
 * - users are unfollowed by default
 * - following a user makes it save as is_followed
 * - using regular save method doesn't un-follow
 * - unfollowing a user makes it save as no longer is_followed
 */
func TestFollowUnfollowUser(t *testing.T) {
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	user := create_dummy_user()
	assert.False(user.IsFollowed)
	err := profile.SaveUser(user)
	assert.NoError(err)

	profile.SetUserFollowed(&user, true)
	assert.True(user.IsFollowed)

	// Ensure the change was persisted
	user_reloaded, err := profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID)  // Verify it's the same user
	assert.True(user_reloaded.IsFollowed)

	err = profile.SaveUser(user)  // should NOT un-set is_followed
	assert.NoError(err)
	user_reloaded, err = profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID)  // Verify it's the same user
	assert.True(user_reloaded.IsFollowed)

	profile.SetUserFollowed(&user, false)
	assert.False(user.IsFollowed)

	// Ensure the change was persisted
	user_reloaded, err = profile.GetUserByHandle(user.Handle)
	require.NoError(t, err)
	assert.Equal(user.ID, user_reloaded.ID)  // Verify it's the same user
	assert.False(user_reloaded.IsFollowed)
}
