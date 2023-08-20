package persistence

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

func (c CursorPosition) IsEnd() bool {
	return c == CURSOR_END
}

// Whether to require, exclude, or indifferent a type of content
type Filter int

const (
	// Filter is not used
	NONE Filter = iota
	// All results must match the filter
	REQUIRE
	// Results must not match the filter
	EXCLUDE
)

type CursorResult struct {
	scraper.Tweet
	scraper.Retweet
	Chrono   int            `db:"chrono"`
	ByUserID scraper.UserID `db:"by_user_id"`
}

type Cursor struct {
	CursorPosition
	CursorValue int
	SortOrder
	PageSize int

	// Search params
	Keywords              []string
	FromUserHandle        scraper.UserHandle
	RetweetedByUserHandle scraper.UserHandle
	ByUserHandle          scraper.UserHandle
	ToUserHandles         []scraper.UserHandle
	SinceTimestamp        scraper.Timestamp
	UntilTimestamp        scraper.Timestamp
	FilterLinks           Filter
	FilterImages          Filter
	FilterVideos          Filter
	FilterPolls           Filter
	FilterSpaces          Filter
	FilterReplies         Filter
	FilterRetweets        Filter
	FilterOfflineFollowed Filter
}

// Generate a cursor with some reasonable defaults
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

		FilterRetweets: EXCLUDE,
	}
}

// Generate a cursor appropriate for fetching the Offline Timeline
func NewTimelineCursor() Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []scraper.UserHandle{},
		SinceTimestamp: scraper.TimestampFromUnix(0),
		UntilTimestamp: scraper.TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		FilterOfflineFollowed: REQUIRE,
	}
}

// Generate a cursor appropriate for fetching a User Feed
func NewUserFeedCursor(h scraper.UserHandle) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []scraper.UserHandle{},
		SinceTimestamp: scraper.TimestampFromUnix(0),
		UntilTimestamp: scraper.TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		ByUserHandle: h,
	}
}

func NewCursorFromSearchQuery(q string) (Cursor, error) {
	ret := NewCursor()
	is_in_quotes := false
	current_token := ""

	for _, char := range q {
		if char == ' ' && !is_in_quotes {
			// Token is finished
			if current_token == "" {
				// Ignore empty tokens
				continue
			}
			// Add the completed token
			if err := ret.apply_token(current_token); err != nil {
				return Cursor{}, err
			}
			current_token = ""
			continue
		}

		if char == '"' {
			if is_in_quotes {
				is_in_quotes = false
				if err := ret.apply_token(current_token); err != nil {
					return Cursor{}, err
				}
				current_token = ""
				continue
			} else {
				is_in_quotes = true
				continue
			}
		}

		// current_token = fmt.Sprintf("%s%s", current_token, char)
		current_token += string(char)
	}

	// End of query string is reached
	if is_in_quotes {
		return Cursor{}, ErrUnmatchedQuotes
	}
	if current_token != "" {
		if err := ret.apply_token(current_token); err != nil {
			return Cursor{}, err
		}
	}
	return ret, nil
}

var ErrInvalidQuery = errors.New("invalid search query")
var ErrUnmatchedQuotes = fmt.Errorf("%w (unmatched quotes)", ErrInvalidQuery)

func (c *Cursor) apply_token(token string) error {
	parts := strings.Split(token, ":")
	if len(parts) < 2 {
		c.Keywords = append(c.Keywords, token)
		return nil
	}
	var err error
	switch parts[0] {
	case "from":
		c.FromUserHandle = scraper.UserHandle(parts[1])
	case "to":
		c.ToUserHandles = append(c.ToUserHandles, scraper.UserHandle(parts[1]))
	case "retweeted_by":
		c.RetweetedByUserHandle = scraper.UserHandle(parts[1])
	case "since":
		c.SinceTimestamp.Time, err = time.Parse("2006-01-02", parts[1])
	case "until":
		c.UntilTimestamp.Time, err = time.Parse("2006-01-02", parts[1])
	case "filter":
		switch parts[1] {
		case "links":
			c.FilterLinks = REQUIRE
		case "images":
			c.FilterImages = REQUIRE
		case "videos":
			c.FilterVideos = REQUIRE
		case "polls":
			c.FilterPolls = REQUIRE
		case "spaces":
			c.FilterSpaces = REQUIRE
		}
	}
	if err != nil {
		return fmt.Errorf("query token %q: %w", token, ErrInvalidQuery)
	}
	return nil
}

func (p Profile) NextPage(c Cursor) (Feed, error) {
	where_clauses := []string{}
	bind_values := []interface{}{}

	// Keywords
	for _, kw := range c.Keywords {
		where_clauses = append(where_clauses, "text like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", kw))
	}

	// From, to, by, and RT'd by user handles
	if c.FromUserHandle != "" {
		where_clauses = append(where_clauses, "user_id = (select id from users where handle like ?)")
		bind_values = append(bind_values, c.FromUserHandle)
	}
	for _, to_user := range c.ToUserHandles {
		where_clauses = append(where_clauses, "reply_mentions like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", to_user))
	}
	if c.RetweetedByUserHandle != "" {
		where_clauses = append(where_clauses, "retweeted_by = (select id from users where handle like ?)")
		bind_values = append(bind_values, c.RetweetedByUserHandle)
	}
	if c.ByUserHandle != "" {
		where_clauses = append(where_clauses, "by_user_id = (select id from users where handle like ?)")
		bind_values = append(bind_values, c.ByUserHandle)
	}

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
	switch c.FilterLinks {
	case REQUIRE:
		where_clauses = append(where_clauses, "exists (select 1 from urls where urls.tweet_id = tweets.id)")
	case EXCLUDE:
		where_clauses = append(where_clauses, "not exists (select 1 from urls where urls.tweet_id = tweets.id)")
	}
	switch c.FilterImages {
	case REQUIRE:
		where_clauses = append(where_clauses, "exists (select 1 from images where images.tweet_id = tweets.id)")
	case EXCLUDE:
		where_clauses = append(where_clauses, "not exists (select 1 from images where images.tweet_id = tweets.id)")
	}
	switch c.FilterVideos {
	case REQUIRE:
		where_clauses = append(where_clauses, "exists (select 1 from videos where videos.tweet_id = tweets.id)")
	case EXCLUDE:
		where_clauses = append(where_clauses, "not exists (select 1 from videos where videos.tweet_id = tweets.id)")
	}
	switch c.FilterPolls {
	case REQUIRE:
		where_clauses = append(where_clauses, "exists (select 1 from polls where polls.tweet_id = tweets.id)")
	case EXCLUDE:
		where_clauses = append(where_clauses, "not exists (select 1 from polls where polls.tweet_id = tweets.id)")
	}
	switch c.FilterSpaces {
	case REQUIRE:
		where_clauses = append(where_clauses, "space_id != 0")
	case EXCLUDE:
		where_clauses = append(where_clauses, "space_id = 0")
	}

	// Filter by lists (e.g., offline-followed)
	switch c.FilterOfflineFollowed {
	case REQUIRE:
		where_clauses = append(where_clauses, "by_user_id in (select id from users where is_followed = 1)")
	case EXCLUDE:
		where_clauses = append(where_clauses, "by_user_id not in (select id from users where is_followed = 1)")
	}
	switch c.FilterReplies {
	case REQUIRE:
		where_clauses = append(where_clauses, "in_reply_to_id != 0")
	case EXCLUDE:
		where_clauses = append(where_clauses, "in_reply_to_id = 0")
	}
	switch c.FilterRetweets {
	case REQUIRE:
		where_clauses = append(where_clauses, "retweet_id != 0")
	case EXCLUDE:
		where_clauses = append(where_clauses, "retweet_id = 0")
	}

	// Pagination
	if c.CursorPosition != CURSOR_START {
		where_clauses = append(where_clauses, c.SortOrder.PaginationWhereClause())
		bind_values = append(bind_values, c.CursorValue)
	}

	where_clause := "where " + strings.Join(where_clauses, " and ")

	q := `select * from (
	select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
           mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
           is_expandable,
           is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
           0 tweet_id, 0 retweet_id, 0 retweeted_by, 0 retweeted_at,
           posted_at chrono, user_id by_user_id
      from tweets
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
     ` + where_clause + ` ` + c.SortOrder.OrderByClause() + ` limit ?
    )

     union

    select * from (
    select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
           mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
           is_expandable,
           is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
           tweet_id, retweet_id, retweeted_by, retweeted_at,
           retweeted_at chrono, retweeted_by by_user_id
      from retweets
 left join tweets on retweets.tweet_id = tweets.id
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
     ` + where_clause + `
     ` + c.SortOrder.OrderByClause() + `
     limit ?
    ) ` + c.SortOrder.OrderByClause() + ` limit ?`

	bind_values = append(bind_values, c.PageSize)
	bind_values = append(bind_values, bind_values...)
	bind_values = append(bind_values, c.PageSize)

	// fmt.Printf("Query: %s\n", q)
	// fmt.Printf("Bind values: %#v\n", bind_values)
	// Run the query
	var results []CursorResult
	err := p.DB.Select(&results, q, bind_values...)
	if err != nil {
		panic(err)
	}

	// Assemble the feed
	ret := NewFeed()
	for _, val := range results {
		// fmt.Printf("\tResult: %#v\n", val)
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
