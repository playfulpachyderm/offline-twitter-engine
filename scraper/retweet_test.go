package scraper_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"offline_twitter/scraper"
)

func TestParseSingleRetweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tweet_that_is_a_retweet.json")
	if err != nil {
		panic(err)
	}
	var api_tweet scraper.APITweet
	err = json.Unmarshal(data, &api_tweet)
	if err != nil {
		t.Errorf(err.Error())
	}

	retweet, err := scraper.ParseSingleRetweet(api_tweet)
	if err != nil {
		t.Errorf(err.Error())
	}

	if retweet.RetweetID != "1404270043018448896" {
		t.Errorf("Expected %q, got %q", "1404270043018448896", retweet.RetweetID)
	}
	if retweet.TweetID != "1404269989646028804" {
		t.Errorf("Expected %q, got %q", "1404269989646028804", retweet.TweetID)
	}
	if retweet.RetweetedByID != "44067298" {
		t.Errorf("Expected %q, got %q", "44067298", retweet.RetweetedBy)
	}
	if retweet.RetweetedAt.Unix() != 1623639042 {
		t.Errorf("Expected %d, got %d", 1623639042, retweet.RetweetedAt.Unix())
	}
}
