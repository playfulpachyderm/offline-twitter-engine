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
		in_reply_to_id scraper.TweetID
		retweeted_status_id scraper.TweetID
		reply_mentions string
	} {
		{"test_responses/single_tweets/tweet_that_is_a_reply_with_gif.json", "", 0, 1395882872729477131, 0, "@michaelmalice"},
		{"test_responses/single_tweets/tweet_with_image.json", "this saddens me every time", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_that_is_a_reply.json", "Noted", 0, 1396194494710788100, 0, "@RvaTeddy @michaelmalice"},
		{"test_responses/single_tweets/tweet_with_4_images.json", "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_at_mentions_in_front.json", "It always does, doesn't it?", 0, 1428907275532476416, 0, "@rob_mose @primalpoly @jmasseypoet @SpaceX"},
		{"test_responses/single_tweets/tweet_with_unicode_chars.json", "The fact that @michaelmalice new book ‘The Anarchist Handbook’ is just absolutely destroying on the charts is the largest white pill I’ve swallowed in years.", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link.json", "", 1422680899670274048, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link2.json", "sometimes they're too dimwitted to even get the wrong title right", 1396194494710788100, 1395882872729477131, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link3.json", "I was using an analogy about creating out-groups but the Germans sure love their literalism", 1442092399358930946, 1335678942020300802, 0, ""},
		{"test_responses/single_tweets/tweet_with_html_entities.json", "By the 1970s  the elite consensus was that \"the hunt for atomic spies\" had been a grotesque over-reaction to minor leaks that cost the lives of the Rosenbergs & ruined many innocents. Only when the USSR fell was it discovered that they & other spies had given away ALL the secrets", 0, 0, 0, ""},
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
		if scraper.TweetID(tweet.InReplyToStatusID) != v.in_reply_to_id {
			t.Errorf("Expected in_reply_to_id id %d, but got %d", v.in_reply_to_id, tweet.InReplyToStatusID)
		}
		if scraper.TweetID(tweet.RetweetedStatusID) != v.retweeted_status_id {
			t.Errorf("Expected retweeted status id %d, but got %d", v.retweeted_status_id, tweet.RetweetedStatusID)
		}
		if tweet.Entities.ReplyMentions != v.reply_mentions {
			t.Errorf("Expected @reply mentions to be %q, but it was %q", v.reply_mentions, tweet.Entities.ReplyMentions)
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


func TestHandleTombstonesHidden(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tombstones/tombstone_hidden_1.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(tweet_resp.GlobalObjects.Tweets) != 2 {
		t.Fatalf("Should have started with %d tweets, but had %d instead", 2, len(tweet_resp.GlobalObjects.Tweets))
	}
	tweet_resp.HandleTombstones()
	if len(tweet_resp.GlobalObjects.Tweets) != 4 {
		t.Errorf("Should have ended up with %d tweets, but had %d instead", 4, len(tweet_resp.GlobalObjects.Tweets))
	}

	first_tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454522147750260742"]
	if !ok {
		t.Errorf("Missing tombstoned tweet for %s", "1454522147750260742")
	}
	if first_tombstone.ID != 1454522147750260742 {
		t.Errorf("Expected ID %d, got %d instead", 1454522147750260742, first_tombstone.ID)
	}
	if first_tombstone.UserID != 1365863538393309184 {
		t.Errorf("Expected UserID %d, got %d instead", 1365863538393309184, first_tombstone.UserID)
	}
	if first_tombstone.TombstoneText != "hidden" {
		t.Errorf("Wrong tombstone text: %s", first_tombstone.TombstoneText)
	}

	second_tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454515503242829830"]
	if !ok {
		t.Errorf("Missing tombstoned tweet for %s", "1454515503242829830")
	}
	if second_tombstone.ID != 1454515503242829830 {
		t.Errorf("Expected ID %d, got %d instead", 1454515503242829830, second_tombstone.ID)
	}
	if second_tombstone.UserID != 1365863538393309184 {
		t.Errorf("Expected UserID %d, got %d instead", 1365863538393309184, second_tombstone.UserID)
	}
	if second_tombstone.TombstoneText != "hidden" {
		t.Errorf("Wrong tombstone text: %s", second_tombstone.TombstoneText)
	}
}

func TestHandleTombstonesDeleted(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tombstones/tombstone_deleted.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(tweet_resp.GlobalObjects.Tweets) != 1 {
		t.Fatalf("Should have started with %d tweets, but had %d instead", 1, len(tweet_resp.GlobalObjects.Tweets))
	}
	tweet_resp.HandleTombstones()
	if len(tweet_resp.GlobalObjects.Tweets) != 2 {
		t.Errorf("Should have ended up with %d tweets, but had %d instead", 2, len(tweet_resp.GlobalObjects.Tweets))
	}

	tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454521654781136902"]
	if !ok {
		t.Errorf("Missing tombstoned tweet for %s", "1454521654781136902")
	}
	if tombstone.ID != 1454521654781136902 {
		t.Errorf("Expected ID %d, got %d instead", 1454521654781136902, tombstone.ID)
	}
	if tombstone.UserID != 1218687933391298560 {
		t.Errorf("Expected UserID %d, got %d instead", 1218687933391298560, tombstone.UserID)
	}
	if tombstone.TombstoneText != "deleted" {
		t.Errorf("Wrong tombstone text: %s", tombstone.TombstoneText)
	}
}

func TestHandleTombstonesUnavailable(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/tombstones/tombstone_unavailable.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp scraper.TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(tweet_resp.GlobalObjects.Tweets) != 2 {
		t.Fatalf("Should have started with %d tweets, but had %d instead", 2, len(tweet_resp.GlobalObjects.Tweets))
	}
	tweet_resp.HandleTombstones()
	if len(tweet_resp.GlobalObjects.Tweets) != 3 {
		t.Errorf("Should have ended up with %d tweets, but had %d instead", 3, len(tweet_resp.GlobalObjects.Tweets))
	}

	tombstone, ok := tweet_resp.GlobalObjects.Tweets["1452686887651532809"]
	if !ok {
		t.Errorf("Missing tombstoned tweet for %s", "1452686887651532809")
	}
	if tombstone.ID != 1452686887651532809 {
		t.Errorf("Expected ID %d, got %d instead", 1452686887651532809, tombstone.ID)
	}
	if tombstone.UserID != 1241389617502445569 {
		t.Errorf("Expected UserID %d, got %d instead", 1241389617502445569, tombstone.UserID)
	}
	if tombstone.TombstoneText != "unavailable" {
		t.Errorf("Wrong tombstone text: %s", tombstone.TombstoneText)
	}
}
