package scraper

type Retweet struct {
	RetweetID     TweetID `db:"retweet_id"`
	TweetID       TweetID `db:"tweet_id"`
	Tweet         *Tweet
	RetweetedByID UserID `db:"retweeted_by"`
	RetweetedBy   *User
	RetweetedAt   Timestamp `db:"retweeted_at"`
}
