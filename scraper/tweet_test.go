package scraper_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"offline_twitter/scraper"
)

func load_tweet_from_file(filename string) scraper.Tweet{
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var apitweet scraper.APITweet
	err = json.Unmarshal(data, &apitweet)
	if err != nil {
		panic(err)
	}
	tweet, err := scraper.ParseSingleTweet(apitweet)
	if err != nil {
		panic(err)
	}
	return tweet
}


func TestParseSingleTweet(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_unicode_chars.json")

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

	if tweet.QuotedTweetID != 0 {
		t.Errorf("Incorrectly believes it quote-tweets tweet with ID %d", tweet.QuotedTweetID)
	}

	if len(tweet.Polls) != 0 {
		t.Errorf("Should not have any polls")
	}
}

func TestParseTweetWithImage(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_image.json")

	expected_text := "this saddens me every time"
	if tweet.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet.Text)
	}
	if len(tweet.Images) != 1 {
		t.Errorf("Expected 1 images but got %d", len(tweet.Images))
	}
}

func TestParseTweetWithQuotedTweetAsLink(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_quoted_tweet_as_link2.json")

	expected_text := "sometimes they're too dimwitted to even get the wrong title right"
	if tweet.Text != expected_text {
		t.Errorf("Expected: %q, got: %q", expected_text, tweet.Text)
	}

	expected_replied_id := scraper.TweetID(1395882872729477131)
	if tweet.InReplyToID != expected_replied_id {
		t.Errorf("Expected %q, got %q", expected_replied_id, tweet.InReplyToID)
	}
	if len(tweet.ReplyMentions) != 0 {
		t.Errorf("Wanted %v, got %v", []string{}, tweet.ReplyMentions)
	}

	expected_quoted_id := scraper.TweetID(1396194494710788100)
	if tweet.QuotedTweetID != expected_quoted_id {
		t.Errorf("Should be a quoted tweet with ID %d, but got %d instead", expected_quoted_id, tweet.QuotedTweetID)
	}

	if len(tweet.Polls) != 0 {
		t.Errorf("Should not have any polls")
	}
}

func TestParseTweetWithVideo(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_video.json")

	expected_video := "https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12"
	if len(tweet.Videos) != 1 || tweet.Videos[0].RemoteURL != expected_video {
		t.Errorf("Expected video URL %q, but got %+v", expected_video, tweet.Videos)
	}
	if tweet.Videos[0].IsGif != false {
		t.Errorf("Expected it to be a regular video, but it was a gif")
	}

	if len(tweet.Images) != 0 {
		t.Errorf("Should not have any images, but has %d", len(tweet.Images))
	}
}

func TestParseTweetWithGif(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_that_is_a_reply_with_gif.json")

	expected_video := "https://video.twimg.com/tweet_video/E189-VhVoAYcrDv.mp4"
	if len(tweet.Videos) != 1 {
		t.Errorf("Expected 1 video (a gif), but got %d instead", len(tweet.Videos))
	}
	if tweet.Videos[0].RemoteURL != expected_video {
		t.Errorf("Expected video URL %q, but got %+v", expected_video, tweet.Videos)
	}
	if tweet.Videos[0].IsGif != true {
		t.Errorf("Expected video to be a gif, but it wasn't")
	}
}

func TestParseTweetWithUrl(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_url_card.json")

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

	if len(tweet.Polls) != 0 {
		t.Errorf("Should not have any polls")
	}
}

func TestParseTweetWithUrlButNoCard(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_url_but_no_card.json")

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
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_multiple_urls.json")

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

	if len(tweet.Polls) != 0 {
		t.Errorf("Should not have any polls")
	}
}

func TestTweetWithLotsOfReplyMentions(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_at_mentions_in_front.json")

	if len(tweet.ReplyMentions) != 4 {
		t.Errorf("Expected %d reply-mentions, got %d", 4, len(tweet.ReplyMentions))
	}
	for i, v := range []scraper.UserHandle{"rob_mose", "primalpoly", "jmasseypoet", "SpaceX"} {
		if tweet.ReplyMentions[i] != v {
			t.Errorf("Expected %q, got %q at position %d", v, tweet.ReplyMentions[i], i)
		}
	}
}

func TestTweetWithPoll(t *testing.T) {
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_poll_4_choices.json")

	if len(tweet.Polls) != 1 {
		t.Fatalf("Expected there to be 1 poll, but there was %d", len(tweet.Polls))
	}
	p := tweet.Polls[0]

	if p.TweetID != tweet.ID {
		t.Errorf("Poll's TweetID (%d) should match the tweet's ID (%d)", p.TweetID, tweet.ID)
	}
	if p.NumChoices != 4 {
		t.Errorf("Expected %d choices, got %d instead", 4, p.NumChoices)
	}
	expected_choice1 := "Tribal armband"
	if p.Choice1 != expected_choice1 {
		t.Errorf("Expected choice1 %q, got %q", expected_choice1, p.Choice1)
	}
	expected_choice2 := "Marijuana leaf"
	if p.Choice2 != expected_choice2 {
		t.Errorf("Expected choice2 %q, got %q", expected_choice2, p.Choice2)
	}
	expected_choice3 := "Butterfly"
	if p.Choice3 != expected_choice3 {
		t.Errorf("Expected choice3 %q, got %q", expected_choice3, p.Choice3)
	}
	expected_choice4 := "Maple leaf"
	if p.Choice4 != expected_choice4 {
		t.Errorf("Expected choice4 %q, got %q", expected_choice4, p.Choice4)
	}

	expected_votes1 := 1593
	expected_votes2 := 624
	expected_votes3 := 778
	expected_votes4 := 1138
	if p.Choice1_Votes != expected_votes1 {
		t.Errorf("Expected Choice1_Votes %d, got %d", expected_votes1, p.Choice1_Votes)
	}
	if p.Choice2_Votes != expected_votes2 {
		t.Errorf("Expected Choice2_Votes %d, got %d", expected_votes2, p.Choice2_Votes)
	}
	if p.Choice3_Votes != expected_votes3 {
		t.Errorf("Expected Choice3_Votes %d, got %d", expected_votes3, p.Choice3_Votes)
	}
	if p.Choice4_Votes != expected_votes4 {
		t.Errorf("Expected Choice4_Votes %d, got %d", expected_votes4, p.Choice4_Votes)
	}

	expected_duration := 1440 * 60
	if p.VotingDuration != expected_duration {
		t.Errorf("Expected voting duration %d seconds, got %d", expected_duration, p.VotingDuration)
	}
	expected_ends_at := int64(1638331934)
	if p.VotingEndsAt.Unix() != expected_ends_at {
		t.Errorf("Expected voting ends at %d (unix), got %d", expected_ends_at, p.VotingEndsAt.Unix())
	}
	expected_last_updated_at := int64(1638331935)
	if p.LastUpdatedAt.Unix() != expected_last_updated_at {
		t.Errorf("Expected updated %d, got %d", expected_last_updated_at, p.LastUpdatedAt.Unix())
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
		t.Errorf("Expected %d retweets, got %d", 3, len(retweets))
	}
	if len(users) != 9 {
		t.Errorf("Expected %d users, got %d", 9, len(users))
	}
}

func TestParseTweetResponseWithTombstones(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tombstones/tombstone_deleted.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Errorf(err.Error())
	}
	extra_users := tweet_resp.HandleTombstones()
	if len(extra_users) != 1 {
		t.Errorf("Expected to need 1 extra user but got %d instead", len(extra_users))
	}

	tweets, retweets, users, err := scraper.ParseTweetResponse(tweet_resp)
	if err != nil {
		t.Fatal(err)
	}

	if len(tweets) != 2 {
		t.Errorf("Expected %d tweets, got %d", 2, len(tweets))
	}
	if len(retweets) != 0 {
		t.Errorf("Expected %d retweets, got %d", 0, len(retweets))
	}
	if len(users) != 1 {
		t.Errorf("Expected %d users, got %d", 1, len(users))
	}
}
