package scraper

import (
	"fmt"
	"errors"
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
	api := API{}
	tweet_response, err := api.GetFeedFor(user_id, "")
	if err != nil {
		err = fmt.Errorf("Error calling API to fetch user feed: UserID %d\n  %w", user_id, err)
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < min_tweets && tweet_response.GetCursor() != "" {
		err = api.GetMoreTweetsFromFeed(user_id, &tweet_response, min_tweets)
		if err != nil && !errors.Is(err, END_OF_FEED) {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}


func GetUserFeedGraphqlFor(user_id UserID, min_tweets int) (trove TweetTrove, err error) {
	api := API{}
	api_response, err := api.GetGraphqlFeedFor(user_id, "")
	if err != nil {
		err = fmt.Errorf("Error calling API to fetch user feed: UserID %d\n  %w", user_id, err)
		return
	}

	if len(api_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries) < min_tweets && api_response.GetCursorBottom() != "" {
		err = api.GetMoreTweetsFromGraphqlFeed(user_id, &api_response, min_tweets)
		if err != nil && !errors.Is(err, END_OF_FEED) {
			return
		}
	}

	trove, err = api_response.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	fmt.Println("------------")
	trove.FetchTombstoneUsers()
	trove.FillMissingUserIDs()

	return trove, nil
}
