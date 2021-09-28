package scraper_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"offline_twitter/scraper"
)


func TestParseSingleTweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_unicode_chars.json")
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

	if tweet.QuotedTweet != 0 {
		t.Errorf("Incorrectly believes it quote-tweets tweet with ID %d", tweet.QuotedTweet)
	}
}

func TestParseTweetWithImage(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_image.json")
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
		t.Fatalf(err.Error())
	}

	expected_text := "this saddens me every time"
	if tweet.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet.Text)
	}
	if len(tweet.Images) != 1 {
		t.Errorf("Expected 1 images but got %d", len(tweet.Images))
	}
}

func TestParseTweetWithQuotedTweetAsLink(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_quoted_tweet_as_link2.json")
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

	expected_text := "sometimes they're too dimwitted to even get the wrong title right"
	if tweet.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet.Text)
	}

	expected_replied_id := scraper.TweetID(1395882872729477131)
	if tweet.InReplyTo != expected_replied_id {
		t.Errorf("Expected %q, got %q", expected_replied_id, tweet.InReplyTo)
	}
	if len(tweet.ReplyMentions) != 1 || tweet.ReplyMentions[0] != "michaelmalice" {
		t.Errorf("Wanted %v, got %v", []string{"michaelmalice"}, tweet.ReplyMentions)
	}

	expected_quoted_id := scraper.TweetID(1396194494710788100)
	if tweet.QuotedTweet != expected_quoted_id {
		t.Errorf("Should be a quoted tweet with ID %d, but got %d instead", expected_quoted_id, tweet.QuotedTweet)
	}
}

func TestParseTweetWithVideo(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_video.json")
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
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_url_card.json")
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
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_url_but_no_card.json")
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

func TestParseTweetWithMultipleUrls(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/single_tweets/tweet_with_multiple_urls.json")
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

	if len(tweet.Urls) != 3 {
		t.Errorf("Expected %d urls, got %d instead", 3, len(tweet.Urls))
	}
	if tweet.Urls[0].HasCard {
		t.Errorf("Expected url not to have a card, but it does: %d", 0)
	}
	if tweet.Urls[1].HasCard {
		t.Errorf("Expected url not to have a card, but it does: %d", 1)
	}
	if !tweet.Urls[2].HasCard {
		t.Errorf("Expected url to have a card, but it doesn't: %d", 2)
	}
	expected_title := "Biden’s victory came from the suburbs"
	if tweet.Urls[2].Title != expected_title {
		t.Errorf("Expected title to be %q, but got %q", expected_title, tweet.Urls[2].Title)
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
