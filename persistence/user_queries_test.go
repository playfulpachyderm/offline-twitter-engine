package persistence_test

import (
	"testing"

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
