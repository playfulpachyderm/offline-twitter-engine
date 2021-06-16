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

	if user.ID != "44067298" {
		t.Errorf("Expected %q, got %q", "44067298", user.ID)
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
	expectedProfileImage := "https://pbs.twimg.com/profile_images/1064051934812913664/Lbwdb_C9_normal.jpg"
	if user.ProfileImageUrl != expectedProfileImage {
		t.Errorf("Expected %q, got %q", expectedProfileImage, user.ProfileImageUrl)
	}
	expectedBannerImage := "https://pbs.twimg.com/profile_banners/44067298/1615134676"
	if user.BannerImageUrl != expectedBannerImage {
		t.Errorf("Expected %q, got %q", expectedBannerImage, user.BannerImageUrl)
	}
	if user.PinnedTweet != scraper.TweetID("1403835414373339136") {
		t.Errorf("Expected %q, got %q", scraper.TweetID("1403835414373339136"), user.PinnedTweet)
	}
}