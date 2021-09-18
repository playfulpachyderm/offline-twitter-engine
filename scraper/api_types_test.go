package scraper_test

import (
	"testing"
	"io/ioutil"
	"encoding/json"

	"offline_twitter/scraper"
)


func TestNormalizeContent(t *testing.T) {
	test_cases := []struct {
		filename string
		eventual_full_text string
		quoted_status_id scraper.TweetID
		in_reply_to scraper.TweetID
		retweeted_status_id scraper.TweetID
	} {
		{"test_responses/tweet_with_gif_reply.json", "", 0, 1395882872729477131, 0},
		{"test_responses/tweet_with_image.json", "this saddens me every time", 0, 0, 0},
		{"test_responses/tweet_with_reply.json", "I always liked \"The Anarchist's Cookbook.\"", 0, 1395978577267593218, 0},
		{"test_responses/tweet_with_4_images.json", "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do", 0, 0, 0},
		{"test_responses/tweet_with_quoted_tweet.json", "", 1422680899670274048, 0, 0},
		{"test_responses/tweet_that_is_a_retweet.json", "RT @nofunin10ded: @michaelmalice We're dealing with people who will napalm your children and then laugh about it", 0, 0, 1404269989646028804},
	}
	for _, v := range test_cases {
		data, err := ioutil.ReadFile(v.filename)
		if err != nil {
			panic(err)
		}
		var tweet scraper.APITweet
		err = json.Unmarshal(data, &tweet)
		if err != nil {
			println("Failed at " + v.filename)
			t.Errorf(err.Error())
		}

		tweet.NormalizeContent()

		if tweet.FullText != v.eventual_full_text {
			t.Errorf("Expected %q, got %q", v.eventual_full_text, tweet.FullText)
		}
		if scraper.TweetID(tweet.QuotedStatusID) != v.quoted_status_id {
			t.Errorf("Expected quoted status %d, but got %d", v.quoted_status_id, tweet.QuotedStatusID)
		}
		if scraper.TweetID(tweet.InReplyToStatusID) != v.in_reply_to {
			t.Errorf("Expected in_reply_to id %d, but got %d", v.in_reply_to, tweet.InReplyToStatusID)
		}
		if scraper.TweetID(tweet.RetweetedStatusID) != v.retweeted_status_id {
			t.Errorf("Expected retweeted status id %d, but got %d", v.retweeted_status_id, tweet.RetweetedStatusID)
		}
	}
}


func TestUserProfileToAPIUser(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/michael_malice_user_profile.json")
	if err != nil {
		panic(err)
	}
	var user_resp scraper.UserResponse
	err = json.Unmarshal(data, &user_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	result := user_resp.ConvertToAPIUser()

	if result.ID != 44067298 {
		t.Errorf("Expected ID %q, got %q", 44067298, result.ID)
	}
	if result.FollowersCount != user_resp.Data.User.Legacy.FollowersCount {
		t.Errorf("Expected user count %d, got %d", user_resp.Data.User.Legacy.FollowersCount, result.FollowersCount)
	}
}


func TestGetCursor(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/midriffs_anarchist_cookbook.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	expected_cursor := "LBmGhsC+ibH1peAmgICjpbS0m98mgICj7a2lmd8mhsC4rbmsmN8mgMCqkbT1p+AmgsC4ucv4o+AmhoCyrf+nlt8mhMC9qfOwlt8mJQISAAA="
	actual_cursor := tweet_resp.GetCursor()

	if expected_cursor != actual_cursor {
		t.Errorf("Expected %q, got %q", expected_cursor, actual_cursor)
	}
}


func TestIsEndOfFeed(t *testing.T) {
	test_cases := []struct {
		filename string
		is_end_of_feed bool
	} {
		{"test_responses/michael_malice_feed.json", false},
		{"test_responses/kwiber_end_of_feed.json", true},
	}
	for _, v := range test_cases {
		data, err := ioutil.ReadFile(v.filename)
		if err != nil {
			panic(err)
		}
		var tweet_resp scraper.TweetResponse
		err = json.Unmarshal(data, &tweet_resp)
		if err != nil {
			t.Fatalf(err.Error())
		}
		result := tweet_resp.IsEndOfFeed()
		if v.is_end_of_feed != result {
			t.Errorf("Expected IsEndOfFeed to be %v, but got %v", v.is_end_of_feed, result)
		}
	}
}
