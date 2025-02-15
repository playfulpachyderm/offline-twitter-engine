package persistence

type NotificationID string

type NotificationType int

const (
	// ActionUserID is who "liked" it, Action[Re]TweetID is most recent [re]tweet they "liked".
	// Can have either many Users "liking" one [re]tweet, or many 1 User "liking" many [re]tweets.
	// The "liked" items can be a mix of Tweets and Retweets.
	NOTIFICATION_TYPE_LIKE NotificationType = iota + 1

	// ActionUserID is who retweeted it, Action[Re]TweetID is your [re]tweet they retweeted.
	// Can have either many Users to one [re]tweet, or many [Re]Tweets from 1 User.
	NOTIFICATION_TYPE_RETWEET

	// ActionUserID is who quote-tweeted you.  ActionTweetID is their tweet.  ActionRetweet is empty
	NOTIFICATION_TYPE_QUOTE_TWEET

	// ActionUserID is who replied to you.  ActionTweetID is their tweet.  ActionRetweet is empty
	NOTIFICATION_TYPE_REPLY

	// ActionUserID is who followed you.  Everything else is empty
	NOTIFICATION_TYPE_FOLLOW

	// ActionTweetID is their tweet.  ActionUserID and ActionRetweetID are empty
	NOTIFICATION_TYPE_MENTION

	// ActionUserID is who is live.  Everything else is empty
	NOTIFICATION_TYPE_USER_IS_LIVE

	// ActionTweetID is the tweet with the poll.  Everything else is empty
	NOTIFICATION_TYPE_POLL_ENDED

	// Everything is empty
	NOTIFICATION_TYPE_LOGIN

	// ActionTweetID is the new pinned post.  Everything else is empty
	NOTIFICATION_TYPE_COMMUNITY_PINNED_POST

	// ActionTweetID is the recommended post.  Everything else is empty
	NOTIFICATION_TYPE_RECOMMENDED_POST
)

type Notification struct {
	ID        NotificationID   `db:"id"`
	Type      NotificationType `db:"type"`
	SentAt    Timestamp        `db:"sent_at"`
	SortIndex int64            `db:"sort_index"`
	UserID    UserID           `db:"user_id"` // recipient of the notification

	ActionUserID    UserID  `db:"action_user_id"`
	ActionTweetID   TweetID `db:"action_tweet_id"`
	ActionRetweetID TweetID `db:"action_retweet_id"`

	// Used for "multiple" notifs, like "user liked multiple tweets"
	HasDetail     bool      `db:"has_detail"`
	LastScrapedAt Timestamp `db:"last_scraped_at"`

	TweetIDs   []TweetID
	UserIDs    []UserID
	RetweetIDs []TweetID
}
