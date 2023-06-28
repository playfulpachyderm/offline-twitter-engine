package scraper

type Retweet struct {
	RetweetID     TweetID `db:"retweet_id"`
	TweetID       TweetID `db:"tweet_id"`
	Tweet         *Tweet
	RetweetedByID UserID `db:"retweeted_by"`
	RetweetedBy   *User
	RetweetedAt   Timestamp `db:"retweeted_at"`
}

func ParseSingleRetweet(apiTweet APITweet) (ret Retweet, err error) {
	apiTweet.NormalizeContent()

	ret.RetweetID = TweetID(apiTweet.ID)
	ret.TweetID = TweetID(apiTweet.RetweetedStatusID)
	ret.RetweetedByID = UserID(apiTweet.UserID)
	ret.RetweetedAt, err = TimestampFromString(apiTweet.CreatedAt)
	if err != nil {
		panic(err)
	}
	return
}
