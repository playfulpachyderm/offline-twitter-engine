package scraper


// Return a list of tweets, including the original and the rest of its thread,
// along with a list of associated users
func GetFeedFull(user_id UserID, max_tweets int) (tweets []Tweet, retweets []Retweet, users []User, err error) {
	api := API{}
	tweet_response, err := api.GetFeedFor(user_id, "")
	if err != nil {
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < max_tweets &&
			tweet_response.GetCursor() != "" {
		err = api.GetMoreTweets(user_id, &tweet_response, max_tweets)
		if err != nil {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}
