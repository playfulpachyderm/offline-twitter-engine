package scraper

import (
	"net/url"
)

func (api API) GetNotifications(cursor string) (TweetResponse, error) {
	url, err := url.Parse("https://api.twitter.com/2/notifications/all.json")
	if err != nil {
		panic(err)
	}

	query := url.Query()
	add_tweet_query_params(&query)
	url.RawQuery = query.Encode()

	var result TweetResponse
	err = api.do_http(url.String(), cursor, &result)

	return result, err
}
