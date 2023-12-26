package scraper

import (
	"errors"
	"fmt"
)

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
func GetUserFeedFor(user_id UserID, min_tweets int) (trove TweetTrove, err error) {
	tweet_response, err := the_api.GetFeedFor(user_id, "")
	if err != nil {
		err = fmt.Errorf("Error calling API to fetch user feed: UserID %d\n  %w", user_id, err)
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < min_tweets && tweet_response.GetCursor() != "" {
		err = the_api.GetMoreTweetsFromFeed(user_id, &tweet_response, min_tweets)
		if err != nil && !errors.Is(err, END_OF_FEED) {
			return
		}
	}

	return tweet_response.ToTweetTrove()
}

func GetUserFeedGraphqlFor(user_id UserID, min_tweets int) (trove TweetTrove, err error) {
	return the_api.GetPaginatedQuery(PaginatedUserFeed{user_id}, min_tweets)
}
