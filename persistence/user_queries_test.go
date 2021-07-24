package persistence_test

import (
	"fmt"
	"testing"
	"time"
	"math/rand"

	"github.com/go-test/deep"

	"offline_twitter/scraper"
	"offline_twitter/persistence"
)

/**
 * Helper function
 */
func create_or_load_profile(profile_path string) persistence.Profile {
	var profile persistence.Profile
	var err error

	if !file_exists(profile_path) {
		profile, err = persistence.NewProfile(profile_path)
	} else {
		profile, err = persistence.LoadProfile(profile_path)
	}
	if err != nil {
		panic(err)
	}
	return profile
}


/**
 * Create a user, save it, reload it, and make sure it comes back the same
 */
func TestSaveAndLoadUser(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	// Generate a new random user ID
	rand.Seed(time.Now().UnixNano())
	userID := fmt.Sprint(rand.Int())

	fake_user := scraper.User{
		ID: scraper.UserID(userID),
		DisplayName: "display name",
		Handle: scraper.UserHandle("handle" + userID),
		Bio: "bio",
		FollowersCount: 0,
		FollowingCount: 1000,
		Location: "location",
		Website:"website",
		JoinDate: time.Now().Truncate(1e9),  // Round to nearest second
		IsVerified: false,
		IsPrivate: true,
		ProfileImageUrl: "profile image url",
		BannerImageUrl: "banner image url",
		PinnedTweetID: scraper.TweetID("234"),
	}

	// Save the user, then reload it and ensure it's the same
	err := profile.SaveUser(fake_user)
	if err != nil {
		panic(err)
	}
	new_fake_user, err := profile.GetUserByID(scraper.UserID(userID))
	if err != nil {
		panic(err)
	}

	if diff := deep.Equal(new_fake_user, fake_user); diff != nil {
		t.Error(diff)
	}

	// Same thing, but get by handle
	new_fake_user2, err := profile.GetUserByHandle(scraper.UserHandle(fake_user.Handle))
	if err != nil {
		panic(err)
	}

	if diff := deep.Equal(new_fake_user2, fake_user); diff != nil {
		t.Error(diff)
	}
}


/**
 * Should correctly report whether the user exists in the database
 */
func TestUserExists(t *testing.T) {
	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	// Generate a new random user ID
	rand.Seed(time.Now().UnixNano())
	userID := fmt.Sprint(rand.Int())

	user := scraper.User{}
	user.ID = scraper.UserID(userID)
	user.Handle = scraper.UserHandle("handle" + userID)

	exists := profile.UserExists(scraper.UserHandle(user.Handle))
	if exists {
		t.Errorf("It shouldn't exist, but it does: %s", userID)
	}
	err := profile.SaveUser(user)
	if err != nil {
		panic(err)
	}
	exists = profile.UserExists(scraper.UserHandle(user.Handle))
	if !exists {
		t.Errorf("It should exist, but it doesn't: %s", userID)
	}
}
