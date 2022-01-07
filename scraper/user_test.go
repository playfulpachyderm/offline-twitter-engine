package scraper_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"offline_twitter/scraper"
)

func TestParseSingleUser(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/michael_malice_user_profile.json")
	if err != nil {
		panic(err)
	}
	var user_resp scraper.UserResponse
	err = json.Unmarshal(data, &user_resp)
	if err != nil {
		t.Errorf(err.Error())
	}
	apiUser := user_resp.ConvertToAPIUser()

	user, err := scraper.ParseSingleUser(apiUser)
	if err != nil {
		t.Errorf(err.Error())
	}

	expected_id := 44067298
	if user.ID != scraper.UserID(expected_id) {
		t.Errorf("Expected %q, got %q", expected_id, user.ID)
	}
	if user.DisplayName != "Michael Malice" {
		t.Errorf("Expected %q, got %q", "Michael Malice", user.DisplayName)
	}
	if user.Handle != "michaelmalice" {
		t.Errorf("Expected %q, got %q", "michaelmalice", user.Handle)
	}
	expectedBio := "Author of Dear Reader, The New Right & The Anarchist Handbook\nHost of \"YOUR WELCOME\" \nSubject of Ego & Hubris by Harvey Pekar\nUnderwear Model\nHe/Him ⚑"
	if user.Bio != expectedBio {
		t.Errorf("Expected %q, got %q", expectedBio, user.Bio)
	}
	if user.FollowingCount != 941 {
		t.Errorf("Expected %d, got %d", 941, user.FollowingCount)
	}
	if user.FollowersCount != 208589 {
		t.Errorf("Expected %d, got %d", 941, user.FollowersCount)
	}
	if user.Location != "Brooklyn" {
		t.Errorf("Expected %q, got %q", "Brooklyn", user.Location)
	}
	if user.Website != "https://amzn.to/3oInafv" {
		t.Errorf("Expected %q, got %q", "https://amzn.to/3oInafv", user.Website)
	}
	if user.JoinDate.Unix() != 1243920952 {
		t.Errorf("Expected %d, got %d", 1243920952, user.JoinDate.Unix())
	}
	if user.IsPrivate != false {
		t.Errorf("Expected %v, got %v", false, user.IsPrivate)
	}
	if user.IsVerified != true {
		t.Errorf("Expected %v, got %v", true, user.IsPrivate)
	}
	if user.IsBanned != false {
		t.Errorf("User should not be banned")
	}
	expectedProfileImage := "https://pbs.twimg.com/profile_images/1064051934812913664/Lbwdb_C9.jpg"
	if user.ProfileImageUrl != expectedProfileImage {
		t.Errorf("Expected %q, got %q", expectedProfileImage, user.ProfileImageUrl)
	}
	expected_tiny_profile_image := "https://pbs.twimg.com/profile_images/1064051934812913664/Lbwdb_C9_normal.jpg"
	if user.GetTinyProfileImageUrl() != expected_tiny_profile_image {
		t.Errorf("Expected %q, got %q", expected_tiny_profile_image, user.GetTinyProfileImageUrl())
	}
	expectedBannerImage := "https://pbs.twimg.com/profile_banners/44067298/1615134676"
	if user.BannerImageUrl != expectedBannerImage {
		t.Errorf("Expected %q, got %q", expectedBannerImage, user.BannerImageUrl)
	}
	expected_profile_image_local := "michaelmalice_profile_Lbwdb_C9.jpg"
	if user.ProfileImageLocalPath != expected_profile_image_local {
		t.Errorf("Expected %q, got %q", expected_profile_image_local, user.ProfileImageLocalPath)
	}
	expected_banner_image_local := "michaelmalice_banner_1615134676.jpg"
	if user.BannerImageLocalPath != expected_banner_image_local {
		t.Errorf("Expected %q, got %q", expected_banner_image_local, user.BannerImageLocalPath)
	}
	expected_id = 1403835414373339136
	if user.PinnedTweetID != scraper.TweetID(expected_id) {
		t.Errorf("Expected %q, got %q", expected_id, user.PinnedTweet)
	}
}

/**
 * Should correctly parse a banned user
 */
func TestParseBannedUser(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/suspended_user.json")
	if err != nil {
		panic(err)
	}
	var user_resp scraper.UserResponse
	err = json.Unmarshal(data, &user_resp)
	if err != nil {
		t.Fatalf(err.Error())
	}
	apiUser := user_resp.ConvertToAPIUser()

	user, err := scraper.ParseSingleUser(apiUser)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if user.ID != 193918550 {
		t.Errorf("Expected id %d, got %d", 193918550, user.ID)
	}
	if user.IsBanned != true {
		t.Errorf("Expected user to be banned")
	}

	// Test generation of profile images for banned user
	if user.GetTinyProfileImageUrl() != "https://abs.twimg.com/sticky/default_profile_images/default_profile.png" {
		t.Errorf("Incorrect tiny profile image URL for banned user: %q", user.GetTinyProfileImageUrl())
	}
	if user.GetTinyProfileImageLocalPath() != "default_profile.png" {
		t.Errorf("Incorrect tiny profile image local path for banned user: %q", user.GetTinyProfileImageLocalPath())
	}
}

/**
 * Should extract a user handle from a tweet URL, or fail if URL is invalid
 */
func TestParseHandleFromTweetUrl(t *testing.T) {
	// Test valid tweet url
	url := "https://twitter.com/kanesays23/status/1429583672827465730"
	expected_user := scraper.UserHandle("kanesays23")
	result, err := scraper.ParseHandleFromTweetUrl(url)
	if err != nil {
		t.Errorf("Unexpected error while parsing handle: %s", err)
	}
	if result != expected_user {
		t.Errorf("Expected handle %q, got %q", expected_user, result)
	}

	// Test invalid url
	_, err = scraper.ParseHandleFromTweetUrl("awjgwekf")
	if err == nil {
		t.Errorf("Should have produced an error for invalid URL")
	}

	// Test empty string
	_, err = scraper.ParseHandleFromTweetUrl("awjgwekf")
	if err == nil {
		t.Errorf("Should have produced an error for invalid URL")
	}
}
