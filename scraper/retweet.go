package scraper

type Retweet struct {
	RetweetID      TweetID
	TweetID        TweetID
	Tweet          *Tweet
	RetweetedByID  UserID
	RetweetedBy    *User
	RetweetedAt    Timestamp
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
