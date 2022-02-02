package scraper

import (
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
func GetUserFeedFor(user_id UserID, min_tweets int) (tweets []Tweet, retweets []Retweet, users []User, err error) {
	api := API{}
	tweet_response, err := api.GetFeedFor(user_id, "")
	if err != nil {
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < min_tweets && tweet_response.GetCursor() != "" {
		err = api.GetMoreTweetsFromFeed(user_id, &tweet_response, min_tweets)
		if err != nil && err != END_OF_FEED {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}


func GetUserFeedGraphqlFor(user_id UserID, min_tweets int) (trove TweetTrove, err error) {
	api := API{}
	api_response, err := api.GetGraphqlFeedFor(user_id, "")
	if err != nil {
		err = fmt.Errorf("Error calling API to fetch user feed: UserID %d\n  %s", user_id, err.Error())
		return
	}

	if len(api_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries) < min_tweets && api_response.GetCursorBottom() != "" {
		err = api.GetMoreTweetsFromGraphqlFeed(user_id, &api_response, min_tweets)
		if err != nil && err != END_OF_FEED {
			return
		}
	}


	trove, err = api_response.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	// DUPE tombstone-user-processing
	fmt.Println("------------")
	for _, handle := range trove.TombstoneUsers {
		fmt.Println(handle)

		user, err := GetUser(handle)
		if err != nil {
			panic(err)
		}
		fmt.Println(user)

		if user.ID == 0 {
			panic(fmt.Sprintf("UserID == 0 (@%s)", handle))
		}

		trove.Users[user.ID] = user
	}
	// Quoted tombstones need their user_id filled out from the tombstoned_users list
	trove.FillMissingUserIDs()

	// <<<<<<< DUPE tombstone-user-processing

	return trove, nil
}
