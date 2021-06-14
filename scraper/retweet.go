package scraper

import (
	"time"
)

type Retweet struct {
	RetweetID TweetID
	TweetID TweetID
	RetweetedBy UserID
	RetweetedAt time.Time
}

func ParseSingleRetweet(apiTweet APITweet) (ret Retweet, err error) {
	ret.RetweetID = TweetID(apiTweet.ID)
	ret.TweetID = TweetID(apiTweet.RetweetedStatusIDStr)
	ret.RetweetedBy = UserID(apiTweet.UserIDStr)
	ret.RetweetedAt, err = time.Parse(time.RubyDate, apiTweet.CreatedAt)
	return
}
