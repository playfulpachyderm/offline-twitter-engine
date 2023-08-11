package persistence

import (
	"fmt"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type SortOrder int

const (
	SORT_ORDER_NEWEST SortOrder = iota
	SORT_ORDER_OLDEST
	SORT_ORDER_MOST_LIKES
	SORT_ORDER_MOST_RETWEETS
)

func (o SortOrder) OrderByClause() string {
	switch o {
	case SORT_ORDER_NEWEST:
		return "order by chrono desc"
	case SORT_ORDER_OLDEST:
		return "order by chrono asc"
	case SORT_ORDER_MOST_LIKES:
		return "order by num_likes desc"
	case SORT_ORDER_MOST_RETWEETS:
		return "order by num_retweets desc"
	default:
		panic(fmt.Sprintf("Invalid sort order: %d", o))
	}
}
func (o SortOrder) PaginationWhereClause() string {
	switch o {
	case SORT_ORDER_NEWEST:
		return "chrono < ?"
	case SORT_ORDER_OLDEST:
		return "chrono > ?"
	case SORT_ORDER_MOST_LIKES:
		return "num_likes < ?"
	case SORT_ORDER_MOST_RETWEETS:
		return "num_retweets < ?"
	default:
		panic(fmt.Sprintf("Invalid sort order: %d", o))
	}
}
func (o SortOrder) NextCursorValue(r CursorResult) int {
	switch o {
	case SORT_ORDER_NEWEST:
		return r.Chrono
	case SORT_ORDER_OLDEST:
		return r.Chrono
	case SORT_ORDER_MOST_LIKES:
		return r.NumLikes
	case SORT_ORDER_MOST_RETWEETS:
		return r.NumRetweets
	default:
		panic(fmt.Sprintf("Invalid sort order: %d", o))
	}
}

// Position in the feed (i.e., whether scrolling up/down is possible)
type CursorPosition int

const (
	// This is the top of the feed; `cursor_position` is invalid;
	CURSOR_START CursorPosition = iota

	// `cursor_position` indicates what should be on the next page;
	CURSOR_MIDDLE

	// Bottom of the feed has been reached.  Subsequent pages will all be empty
	CURSOR_END
)

type CursorResult struct {
	scraper.Tweet
	scraper.Retweet
	Chrono int `db:"chrono"`
}

type Cursor struct {
	CursorPosition
	CursorValue int
	SortOrder
	PageSize int

	// Search params
	Keywords              []string
	FromUserHandle        scraper.UserHandle
	ToUserHandles         []scraper.UserHandle
	RetweetedByUserHandle scraper.UserHandle
	SinceTimestamp        scraper.Timestamp
	UntilTimestamp        scraper.Timestamp
	FilterLinks           bool
	FilterImages          bool
	FilterVideos          bool
	FilterPolls           bool
}

func NewCursor() Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []scraper.UserHandle{},
		SinceTimestamp: scraper.TimestampFromUnix(0),
		UntilTimestamp: scraper.TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,
	}
}

func (p Profile) NextPage(c Cursor) (Feed, error) {
	where_clauses := []string{}
	bind_values := []interface{}{}

	// Keywords
	for _, kw := range c.Keywords {
		where_clauses = append(where_clauses, "text like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", kw))
	}

	// From, to, and RT'd by user handles
	if c.FromUserHandle != "" {
		where_clauses = append(where_clauses, "user_id = (select id from users where handle like ?)")
		bind_values = append(bind_values, c.FromUserHandle)
	}
	for _, to_user := range c.ToUserHandles {
		where_clauses = append(where_clauses, "reply_mentions like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", to_user))
	}
	where_clauses = append(where_clauses, "retweeted_by = coalesce((select id from users where handle like ?), 0)")
	bind_values = append(bind_values, c.RetweetedByUserHandle)

	// Since and until timestamps
	if c.SinceTimestamp.Unix() != 0 {
		where_clauses = append(where_clauses, "posted_at > ?")
		bind_values = append(bind_values, c.SinceTimestamp)
	}
	if c.UntilTimestamp.Unix() != 0 {
		where_clauses = append(where_clauses, "posted_at < ?")
		bind_values = append(bind_values, c.UntilTimestamp)
	}

	// Media filters
	if c.FilterLinks {
		where_clauses = append(where_clauses, "exists (select 1 from urls where urls.tweet_id = tweets.id)")
	}
	if c.FilterImages {
		where_clauses = append(where_clauses, "exists (select 1 from images where images.tweet_id = tweets.id)")
	}
	if c.FilterVideos {
		where_clauses = append(where_clauses, "exists (select 1 from videos where videos.tweet_id = tweets.id)")
	}
	if c.FilterPolls {
		where_clauses = append(where_clauses, "exists (select 1 from polls where polls.tweet_id = tweets.id)")
	}

	// Pagination
	if c.CursorPosition != CURSOR_START {
		where_clauses = append(where_clauses, c.SortOrder.PaginationWhereClause())
		bind_values = append(bind_values, c.CursorValue)
	}

	where_clause := "where " + strings.Join(where_clauses, " and ")

	q := `select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
           mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
           is_expandable,
           is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
           0 tweet_id, 0 retweet_id, 0 retweeted_by, 0 retweeted_at,
           posted_at chrono
      from tweets
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
     ` + where_clause + `

     union

    select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
           mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
           is_expandable,
           is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
           tweet_id, retweet_id, retweeted_by, retweeted_at,
           retweeted_at chrono
      from retweets
 left join tweets on retweets.tweet_id = tweets.id
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
     ` + where_clause + `
     ` + c.SortOrder.OrderByClause() + `
     limit ?`

	bind_values = append(bind_values, bind_values...)
	bind_values = append(bind_values, c.PageSize)

	// Run the query
	var results []CursorResult
	err := p.DB.Select(&results, q, bind_values...)
	if err != nil {
		panic(err)
	}

	// Assemble the feed
	ret := NewFeed()
	for _, val := range results {
		ret.Tweets[val.Tweet.ID] = val.Tweet
		if val.Retweet.RetweetID != 0 {
			ret.Retweets[val.Retweet.RetweetID] = val.Retweet
		}
		ret.Items = append(ret.Items, FeedItem{TweetID: val.Tweet.ID, RetweetID: val.Retweet.RetweetID})
	}

	p.fill_content(&ret.TweetTrove)

	ret.CursorBottom = c

	// Set the new cursor position and value
	if len(results) < c.PageSize {
		ret.CursorBottom.CursorPosition = CURSOR_END
	} else {
		ret.CursorBottom.CursorPosition = CURSOR_MIDDLE
		last_item := results[len(results)-1]
		ret.CursorBottom.CursorValue = c.SortOrder.NextCursorValue(last_item)
	}

	return ret, nil
}
