package persistence

import (
	"database/sql/driver"
	"errors"
	"strings"
)

var ERR_NO_TWEET = errors.New("Empty tweet")

type TweetID int64

type CommaSeparatedList []string

func (l *CommaSeparatedList) Scan(src interface{}) error {
	*l = CommaSeparatedList{}
	switch src := src.(type) {
	case string:
		for _, v := range strings.Split(src, ",") {
			if v != "" {
				*l = append(*l, v)
			}
		}
	default:
		panic("Should be a string")
	}
	return nil
}
func (l CommaSeparatedList) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}

type Tweet struct {
	ID             TweetID   `db:"id"`
	Text           string    `db:"text"`
	IsExpandable   bool      `db:"is_expandable"`
	PostedAt       Timestamp `db:"posted_at"`
	NumLikes       int       `db:"num_likes"`
	NumRetweets    int       `db:"num_retweets"`
	NumReplies     int       `db:"num_replies"`
	NumQuoteTweets int       `db:"num_quote_tweets"`
	InReplyToID    TweetID   `db:"in_reply_to_id"`
	QuotedTweetID  TweetID   `db:"quoted_tweet_id"`

	UserID UserID `db:"user_id"`
	User   *User  `db:"user"`

	// For processing tombstones
	UserHandle          UserHandle
	InReplyToUserHandle UserHandle
	InReplyToUserID     UserID

	Images        []Image
	Videos        []Video
	Urls          []Url
	Polls         []Poll
	Mentions      CommaSeparatedList `db:"mentions"`
	ReplyMentions CommaSeparatedList `db:"reply_mentions"`
	Hashtags      CommaSeparatedList `db:"hashtags"`

	// TODO get-rid-of-redundant-spaces: Might be good to get rid of `Spaces`.  Only used in APIv1 I think.
	// A first-step would be to delete the Spaces after pulling them out of a Tweet into the Trove
	// in ToTweetTrove.  Then they will only be getting saved once rather than twice.
	Spaces  []Space
	SpaceID SpaceID `db:"space_id"`

	TombstoneType string `db:"tombstone_type"`
	TombstoneText string `db:"tombstone_text"`
	IsStub        bool   `db:"is_stub"`

	IsLikedByCurrentUser  bool      `db:"is_liked_by_current_user"`
	IsContentDownloaded   bool      `db:"is_content_downloaded"`
	IsConversationScraped bool      `db:"is_conversation_scraped"`
	LastScrapedAt         Timestamp `db:"last_scraped_at"`
}
