package scraper

import (
	"errors"
)

func TimestampToDateString(timestamp int) string {
	panic("???")  // TODO
}

/**
 * TODO: Search modes:
 * - regular ("top")
 * - latest / "live"
 * - search for users
 * - photos
 * - videos
 */
func Search(query string, min_results int) (trove TweetTrove, err error) {
	api := API{}
	tweet_response, err := api.Search(query, "")
	if err != nil {
		return
	}

	if len(tweet_response.GlobalObjects.Tweets) < min_results && tweet_response.GetCursor() != "" {
		err = api.GetMoreTweetsFromSearch(query, &tweet_response, min_results)
		if errors.Is(err, END_OF_FEED) {
			println("End of feed!")
		} else if err != nil {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}
