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
	expected_id := 1404270043018448896
	if retweet.RetweetID != scraper.TweetID(1404270043018448896) {
		t.Errorf("Expected %d, got %d", expected_id, retweet.RetweetID)
	}
	expected_id = 1404269989646028804
	if retweet.TweetID != scraper.TweetID(expected_id) {
		t.Errorf("Expected %d, got %d", expected_id, retweet.TweetID)
	}
	expected_id = 44067298
	if retweet.RetweetedByID != scraper.UserID(expected_id) {
		t.Errorf("Expected %d, got %d", expected_id, retweet.RetweetedByID)
	}
	expected_id = 1623639042
	if retweet.RetweetedAt.Unix() != int64(expected_id) {
		t.Errorf("Expected %d, got %d", expected_id, retweet.RetweetedAt.Unix())
	}
}
