//go:build obsolete_user_feed

// Nothing in this file is used.  It's outdated; user feed comes from APIv2 instead now.

package scraper

import (
	"errors"
	"fmt"
	"net/url"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

func (api API) GetFeedFor(user_id UserID, cursor string) (APIv1Response, error) {
	url, err := url.Parse(fmt.Sprintf("%s%d.json", API_USER_TIMELINE_BASE_PATH, user_id))
	if err != nil {
		panic(err)
	}
	queryParams := url.Query()
	add_tweet_query_params(&queryParams)
	url.RawQuery = queryParams.Encode()

	var result APIv1Response
	err = api.do_http(url.String(), cursor, &result)

	return result, err
}

/**
 * Resend the request to get more tweets if necessary
 *
 * args:
 * - user_id: the user's UserID
 * - response: an "out" parameter; the APIv1Response that tweets, RTs and users will be appended to
 * - min_tweets: the desired minimum amount of tweets to get
 */
func (api API) GetMoreTweetsFromFeed(user_id UserID, response *APIv1Response, min_tweets int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < min_tweets {
		fresh_response, err := api.GetFeedFor(user_id, last_response.GetCursor())
		if err != nil {
			return err
		}

		if fresh_response.GetCursor() == last_response.GetCursor() && len(fresh_response.GlobalObjects.Tweets) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			return END_OF_FEED
		}
		if fresh_response.IsEndOfFeed() {
			// Response has a pinned tweet, but no other content: end of feed has been reached
			return END_OF_FEED
		}

		last_response = &fresh_response

		// Copy over the tweets and the users
		for id, tweet := range last_response.GlobalObjects.Tweets {
			response.GlobalObjects.Tweets[id] = tweet
		}
		for id, user := range last_response.GlobalObjects.Users {
			response.GlobalObjects.Users[id] = user
		}
		fmt.Printf("Have %d tweets, and %d users so far\n", len(response.GlobalObjects.Tweets), len(response.GlobalObjects.Users))
	}
	return nil
}

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

/**
 * Return a list of tweets, including the original and the rest of its thread,
 * along with a list of associated users.
 *
 * Mark the main tweet as "is_conversation_downloaded = true", and update its "last_scraped_at"
 * value.
 *
 * args:
 * - id: the ID of the tweet to get
 *
 * returns: the tweet, list of its replies and context, and users associated with those replies
 */
func GetTweetFull(id TweetID, how_many int) (trove TweetTrove, err error) {
	tweet_response, err := the_api.GetTweet(id, "")
	if err != nil {
		err = fmt.Errorf("Error getting tweet: %d\n  %w", id, err)
		return
	}
	if len(tweet_response.GlobalObjects.Tweets) < how_many &&
		tweet_response.GetCursor() != "" {
		err = the_api.GetMoreReplies(id, &tweet_response, how_many)
		if err != nil {
			err = fmt.Errorf("Error getting more tweet replies: %d\n  %w", id, err)
			return
		}
	}

	// This has to be called BEFORE ToTweetTrove, because it modifies the APIv1Response (adds tombstone tweets to its tweets list)
	tombstoned_users := tweet_response.HandleTombstones()

	trove, err = tweet_response.ToTweetTrove()
	if err != nil {
		panic(err)
	}
	trove.TombstoneUsers = tombstoned_users

	// Quoted tombstones need their user_id filled out from the tombstoned_users list
	log.Debug("Running tweet trove post-processing\n")
	err = trove.PostProcess()
	if err != nil {
		err = fmt.Errorf("Error getting tweet (id %d):\n  %w", id, err)
		return
	}

	// Find the main tweet and update its "is_conversation_downloaded" and "last_scraped_at"
	tweet, ok := trove.Tweets[id]
	if !ok {
		panic("Trove didn't contain its own tweet!")
	}
	tweet.LastScrapedAt = Timestamp{time.Now()}
	tweet.IsConversationScraped = true
	trove.Tweets[id] = tweet

	return
}

func (api *API) GetTweet(id TweetID, cursor string) (APIv1Response, error) {
	url, err := url.Parse(fmt.Sprintf("%s%d.json", API_CONVERSATION_BASE_PATH, id))
	if err != nil {
		panic(err)
	}
	queryParams := url.Query()
	if cursor != "" {
		queryParams.Add("referrer", "tweet")
	}
	add_tweet_query_params(&queryParams)
	url.RawQuery = queryParams.Encode()

	var result APIv1Response
	err = api.do_http(url.String(), cursor, &result)
	return result, err
}

// Resend the request to get more replies if necessary
func (api *API) GetMoreReplies(tweet_id TweetID, response *APIv1Response, max_replies int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < max_replies {
		fresh_response, err := api.GetTweet(tweet_id, last_response.GetCursor())
		if err != nil {
			return err
		}

		last_response = &fresh_response

		// Copy over the tweets and the users
		for id, tweet := range last_response.GlobalObjects.Tweets {
			response.GlobalObjects.Tweets[id] = tweet
		}
		for id, user := range last_response.GlobalObjects.Users {
			response.GlobalObjects.Users[id] = user
		}
	}
	return nil
}
