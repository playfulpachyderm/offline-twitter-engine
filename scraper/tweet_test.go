package scraper_test

import (
	"fmt"
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

	if len(tweet.Urls) != 0 {
		t.Errorf("Expected %d urls, but got %d", 0, len(tweet.Urls))
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
	if tweet1.QuotedTweet != 0 {
		t.Errorf("Incorrectly believes it quote-tweets %q", tweet1.QuotedTweet)
	}

	if tweet2.QuotedTweet == 0 {
		t.Errorf("Should be a quoted tweet")
	}

	quoted_tweet_, ok := tweets[fmt.Sprint(tweet2.QuotedTweet)]
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


func TestParseTweetWithVideo(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tweet_with_video.json")
	if err != nil {
		panic(err)
	}
	var apitweet scraper.APITweet
	err = json.Unmarshal(data, &apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}
	tweet, err := scraper.ParseSingleTweet(apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}
	expected_video := "https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12"
	if len(tweet.Videos) != 1 || tweet.Videos[0].RemoteURL != expected_video {
		t.Errorf("Expected video URL %q, but got %+v", expected_video, tweet.Videos)
	}

	if len(tweet.Images) != 0 {
		t.Errorf("Should not have any images, but has %d", len(tweet.Images))
	}
}

func TestParseTweetWithUrl(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tweet_with_url_card.json")
	if err != nil {
		panic(err)
	}
	var apitweet scraper.APITweet
	err = json.Unmarshal(data, &apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}
	tweet, err := scraper.ParseSingleTweet(apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(tweet.Urls) != 1 {
		t.Errorf("Expected %d urls, but got %d", 1, len(tweet.Urls))
	}

	expected_url_text := "https://reason.com/2021/08/30/la-teachers-union-cecily-myart-cruz-learning-loss/"
	if tweet.Urls[0].Text != expected_url_text {
		t.Errorf("Expected Url text to be %q, but got %q", expected_url_text, tweet.Urls[0].Text)
	}
	if !tweet.Urls[0].HasCard {
		t.Errorf("Expected it to have a card, but it doesn't")
	}
	expected_url_domain := "reason.com"
	if tweet.Urls[0].Domain != expected_url_domain {
		t.Errorf("Expected Url text to be %q, but got %q", expected_url_domain, tweet.Urls[0].Domain)
	}
}

func TestParseTweetWithUrlButNoCard(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tweet_with_url_but_no_card.json")
	if err != nil {
		panic(err)
	}
	var apitweet scraper.APITweet
	err = json.Unmarshal(data, &apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}
	tweet, err := scraper.ParseSingleTweet(apitweet)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(tweet.Urls) != 1 {
		t.Errorf("Expected %d urls, but got %d", 1, len(tweet.Urls))
	}

	expected_url_text := "https://www.politico.com/newsletters/west-wing-playbook/2021/09/16/the-jennifer-rubin-wh-symbiosis-494364"
	if tweet.Urls[0].Text != expected_url_text {
		t.Errorf("Expected Url text to be %q, but got %q", expected_url_text, tweet.Urls[0].Text)
	}
	if tweet.Urls[0].HasCard {
		t.Errorf("Expected url not to have a card, but it thinks it has one")
	}
}

func TestParseTweetResponse(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/michael_malice_feed.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweets, retweets, users, err := scraper.ParseTweetResponse(tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(tweets) != 29 - 3 {
		t.Errorf("Expected %d tweets, got %d", 29-3, len(tweets))
	}
	if len(retweets) != 3 {
		t.Errorf("Expected %d tweets, got %d", 3, len(retweets))
	}
	if len(users) != 9 {
		t.Errorf("Expected %d tweets, got %d", 9, len(users))
	}
}
