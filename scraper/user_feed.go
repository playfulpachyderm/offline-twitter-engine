package scraper


/**
 * Get a list of tweets that appear on the given user's page, along with a list of associated
 * users for any retweets.
 *
 * args:
 * - user_id: the ID of the user whomst feed to fetch
 * - min_tweets: get at least this many tweets, if there are any
 *
 * returns: a slice of Tweets, Retweets, and Users
 */
func GetUserFeedFor(user_id UserID, min_tweets int) (tweets []Tweet, retweets []Retweet, users []User, err error) {
	api := API{}
	tweet_response, err := api.GetFeedFor(user_id, "")
	if err != nil {
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < min_tweets &&
			tweet_response.GetCursor() != "" {
		err = api.GetMoreTweetsFromFeed(user_id, &tweet_response, min_tweets)
		if err != nil && err != END_OF_FEED {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}
