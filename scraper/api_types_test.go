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
	} {
		{"test_responses/tweet_with_gif_reply.json", ""},
		{"test_responses/tweet_with_image.json", "this saddens me every time"},
		{"test_responses/tweet_with_reply.json", "I always liked \"The Anarchist's Cookbook.\""},
	}
	for _, v := range test_cases {
		data, err := ioutil.ReadFile(v.filename)
		if err != nil {
			panic(err)
		}
		var tweet scraper.APITweet
		err = json.Unmarshal(data, &tweet)
		if err != nil {
			t.Errorf(err.Error())
		}

		tweet.NormalizeContent()

		if tweet.FullText != v.eventual_full_text {
			t.Errorf("Expected %q, got %q", v.eventual_full_text, tweet.FullText)
		}
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
