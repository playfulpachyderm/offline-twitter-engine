package scraper_test

import (
	// "fmt"
	"encoding/json"
	"io/ioutil"
	"testing"

	"offline_twitter/scraper"
)

func TestParseSingleTweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/dave_smith_anarchist_handbook.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweets := tweet_resp.GlobalObjects.Tweets
	users := tweet_resp.GlobalObjects.Users

	if len(tweets) != 11 {
		t.Errorf("Expected %d tweets, got %d instead", 11, len(tweets))
	}

	if len(users) != 11 {
		t.Errorf("Expected %d users, got %d instead", 11, len(users))
	}

	dave_smith_tweet, ok := tweets["1395881699142160387"]
	if !ok {
		t.Errorf("Didn't find the Dave Smith tweet.")
	}

	tweet, err := scraper.ParseSingleTweet(dave_smith_tweet)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected_text := "The fact that @michaelmalice new book ‘The Anarchist Handbook’ is just absolutely destroying on the charts is the largest white pill I’ve swallowed in years."
	actual_text := tweet.Text

	if actual_text != expected_text {
		t.Errorf("Expected: %q; got %q", expected_text, actual_text)
	}

	if len(tweet.Mentions) != 1 || tweet.Mentions[0] != "michaelmalice" {
		t.Errorf("Expected %v, got %v", []string{"michaelmalice"}, tweet.Mentions)
	}

	if tweet.PostedAt.Unix() != 1621639105 {
		t.Errorf("Expected %d, got %d", 1621639105, tweet.PostedAt.Unix())
	}
}

func TestParseSingleTweet2(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/midriffs_anarchist_cookbook.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweets := tweet_resp.GlobalObjects.Tweets
	users := tweet_resp.GlobalObjects.Users

	if len(tweets) != 12 {
		t.Errorf("Expected %d tweets, got %d instead", 11, len(tweets))
	}

	if len(users) != 11 {
		t.Errorf("Expected %d users, got %d instead", 11, len(users))
	}

	t1, ok := tweets["1395882872729477131"]
	if !ok {
		t.Fatalf("Didn't find first tweet")
	}
	t2, ok := tweets["1396194922009661441"]
	if !ok {
		t.Fatalf("Didn't find second tweet")
	}

	tweet1, err := scraper.ParseSingleTweet(t1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	tweet2, err := scraper.ParseSingleTweet(t2)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected_text := "this saddens me every time"
	if tweet1.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet1.Text)
	}
	expected_text = "sometimes they're too dimwitted to even get the wrong title right"
	if tweet2.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet2.Text)
	}

	if len(tweet1.Images) != 1 {
		t.Errorf("Expected 1 images but got %d", len(tweet1.Images))
	}

	if tweet2.InReplyTo != tweet1.ID {
		t.Errorf("Expected %q, got %q", tweet1.ID, tweet2.InReplyTo)
	}
	if tweet1.QuotedTweet != "" {
		t.Errorf("Incorrectly believes it quote-tweets %q", tweet1.QuotedTweet)
	}

	if tweet2.QuotedTweet == "" {
		t.Errorf("Should be a quoted tweet")
	}

	quoted_tweet_, ok := tweets[string(tweet2.QuotedTweet)]
	if !ok {
		t.Errorf("Couldn't find the quoted tweet")
	}

	quoted_tweet, err := scraper.ParseSingleTweet(quoted_tweet_)
	if err != nil {
		t.Errorf(err.Error())
	}

	expected_text = "I always liked \"The Anarchist's Cookbook.\""
	if quoted_tweet.Text != expected_text {
		t.Errorf("Expected %q, got %q", expected_text, quoted_tweet.Text)
	}
}
