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
	ID        NotificationID `db:"id"`
	Type      int            `db:"type"`
	SentAt    Timestamp      `db:"sent_at"`
	SortIndex int64          `db:"sort_index"`
	UserID    UserID         `db:"user_id"` // recipient of the notification

	ActionUserID    UserID  `db:"action_user_id"`
	ActionTweetID   TweetID `db:"action_tweet_id"`
	ActionRetweetID TweetID `db:"action_retweet_id"`

	TweetIDs []TweetID
	UserIDs  []UserID
}
