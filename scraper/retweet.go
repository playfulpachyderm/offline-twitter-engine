package scraper

import (
	"time"
)

type Retweet struct {
	RetweetID      TweetID
	TweetID        TweetID
	Tweet          *Tweet
	RetweetedByID  UserID
	RetweetedBy    *User
	RetweetedAt    time.Time
}

func ParseSingleRetweet(apiTweet APITweet) (ret Retweet, err error) {
	apiTweet.NormalizeContent()

	ret.RetweetID = TweetID(apiTweet.ID)
	ret.TweetID = TweetID(apiTweet.RetweetedStatusID)
	ret.RetweetedByID = UserID(apiTweet.UserID)
	ret.RetweetedAt, err = time.Parse(time.RubyDate, apiTweet.CreatedAt)
	return
}
