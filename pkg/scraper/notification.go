package scraper

type NotificationID string

type Notification struct {
	ID        NotificationID
	Type      int
	SentAt    Timestamp
	SortIndex int64
	UserID    UserID // recipient of the notification

	ActionUserID    UserID
	ActionTweetID   TweetID
	ActionRetweetID TweetID

	TweetIDs []TweetID
	UserIDs  []UserID
}
