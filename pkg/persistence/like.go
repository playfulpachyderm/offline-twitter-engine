package persistence

type LikeSortID int64

type Like struct {
	SortID  LikeSortID `db:"sort_order"`
	UserID  UserID     `db:"user_id"`
	TweetID TweetID    `db:"tweet_id"`
}
