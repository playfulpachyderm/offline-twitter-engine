package scraper

type BookmarkSortID int64

type Bookmark struct {
	SortID  BookmarkSortID `db:"sort_order"`
	UserID  UserID         `db:"user_id"`
	TweetID TweetID        `db:"tweet_id"`
}
