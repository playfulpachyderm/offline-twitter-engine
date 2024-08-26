package scraper

type NotificationID string

type NotificationType int

const (
	NOTIFICATION_TYPE_LIKE = iota + 1
	NOTIFICATION_TYPE_RETWEET
	NOTIFICATION_TYPE_QUOTE_TWEET
	NOTIFICATION_TYPE_REPLY
	NOTIFICATION_TYPE_FOLLOW
	NOTIFICATION_TYPE_MENTION
	NOTIFICATION_TYPE_USER_IS_LIVE
	NOTIFICATION_TYPE_POLL_ENDED
	NOTIFICATION_TYPE_LOGIN
	NOTIFICATION_TYPE_COMMUNITY_PINNED_POST
	NOTIFICATION_TYPE_RECOMMENDED_POST
)

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
