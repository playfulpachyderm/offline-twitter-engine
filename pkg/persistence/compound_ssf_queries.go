package persistence

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SortOrder int

const (
	SORT_ORDER_NEWEST SortOrder = iota
	SORT_ORDER_OLDEST
	SORT_ORDER_MOST_LIKES
	SORT_ORDER_MOST_RETWEETS
	SORT_ORDER_LIKED_AT
	SORT_ORDER_BOOKMARKED_AT
)

func (o SortOrder) String() string {
	return []string{"newest", "oldest", "most likes", "most retweets", "liked at", "bookmarked at"}[o]
}

func SortOrderFromString(s string) (SortOrder, bool) {
	result, is_ok := map[string]SortOrder{
		"newest":        SORT_ORDER_NEWEST,
		"oldest":        SORT_ORDER_OLDEST,
		"most likes":    SORT_ORDER_MOST_LIKES,
		"most retweets": SORT_ORDER_MOST_RETWEETS,
		"liked at":      SORT_ORDER_LIKED_AT,
		"bookmarked at": SORT_ORDER_BOOKMARKED_AT,
	}[s]
	return result, is_ok // Have to store as temporary variable b/c otherwise it interprets it as single-value and compile fails
}

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
	case SORT_ORDER_LIKED_AT:
		return "order by likes_sort_order desc"
	case SORT_ORDER_BOOKMARKED_AT:
		return "order by bookmarks_sort_order desc"
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
	case SORT_ORDER_LIKED_AT:
		return "likes_sort_order < ?"
	case SORT_ORDER_BOOKMARKED_AT:
		return "bookmarks_sort_order < ?"
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
	case SORT_ORDER_LIKED_AT:
		return r.LikeSortOrder
	case SORT_ORDER_BOOKMARKED_AT:
		return r.BookmarkSortOrder
	default:
		panic(fmt.Sprintf("Invalid sort order: %d", o))
	}
}
func (o SortOrder) NextDMCursorValue(m DMMessage) int64 {
	switch o {
	case SORT_ORDER_NEWEST, SORT_ORDER_OLDEST:
		return m.SentAt.UnixMilli()
	default:
		panic(fmt.Sprintf("Invalid sort order for DMs: %d", o))
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
	Tweet
	Retweet
	Chrono            int    `db:"chrono"`
	LikeSortOrder     int    `db:"likes_sort_order"`
	BookmarkSortOrder int    `db:"bookmarks_sort_order"`
	ByUserID          UserID `db:"by_user_id"`
}

type Cursor struct {
	CursorPosition
	CursorValue int
	SortOrder
	PageSize int

	// Search params
	Keywords               []string
	FromUserHandle         UserHandle   // Tweeted by this user
	RetweetedByUserHandle  UserHandle   // Retweeted by this user
	ByUserHandle           UserHandle   // Either tweeted or retweeted by this user
	ToUserHandles          []UserHandle // In reply to these users
	LikedByUserHandle      UserHandle   // Liked by this user
	BookmarkedByUserHandle UserHandle   // Bookmarked by this user
	ListID                 ListID       // Either tweeted or retweeted by users from this List
	FollowedByUserHandle   UserHandle   // Either tweeted or retweeted by users followed by this user
	SinceTimestamp         Timestamp
	UntilTimestamp         Timestamp
	TombstoneType          string
	FilterLinks            Filter
	FilterImages           Filter
	FilterVideos           Filter
	FilterMedia            Filter
	FilterPolls            Filter
	FilterSpaces           Filter
	FilterReplies          Filter
	FilterRetweets         Filter
	FilterOfflineFollowed  Filter
	QuotedTweetID          TweetID
}

// Generate a cursor with some reasonable defaults
func NewCursor() Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
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
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		FilterOfflineFollowed: REQUIRE,
	}
}

// Generate a cursor appropriate for showing a List feed
func NewListCursor(list_id ListID) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		ListID:         list_id,
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,
	}
}

// Generate a cursor appropriate for fetching a User Feed
func NewUserFeedCursor(h UserHandle) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		ByUserHandle: h,
	}
}

// Generate a cursor appropriate for a user's Media tab
func NewUserFeedMediaCursor(h UserHandle) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		ByUserHandle: h,
		FilterMedia:  REQUIRE,
	}
}

// Generate a cursor for a User's Likes
func NewUserFeedLikesCursor(h UserHandle) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_LIKED_AT,
		PageSize:       50,

		LikedByUserHandle: h,
	}
}

// Generate a cursor for a User's Bookmarks
func NewUserFeedBookmarksCursor(h UserHandle) Cursor {
	return Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_BOOKMARKED_AT,
		PageSize:       50,

		BookmarkedByUserHandle: h,
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
		c.FromUserHandle = UserHandle(parts[1])
	case "to":
		c.ToUserHandles = append(c.ToUserHandles, UserHandle(parts[1]))
	case "retweeted_by":
		c.RetweetedByUserHandle = UserHandle(parts[1])
		c.FilterRetweets = NONE // Clear the "exclude retweets" filter set by default in NewCursor
	case "liked_by":
		c.LikedByUserHandle = UserHandle(parts[1])
	case "bookmarked_by":
		c.BookmarkedByUserHandle = UserHandle(parts[1])
	case "followed_by":
		c.FollowedByUserHandle = UserHandle(parts[1])
	case "quoted_tweet_id":
		t_id, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("%w: filter 'quoted_tweet_id:' must be a number, got %q", ErrInvalidQuery, parts[1])
		}
		c.QuotedTweetID = TweetID(t_id)
	case "list":
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("%w: filter 'list:' must be a number (list ID), got %q", ErrInvalidQuery, parts[1])
		}
		c.ListID = ListID(i)
	case "since":
		c.SinceTimestamp.Time, err = time.Parse("2006-01-02", parts[1])
	case "until":
		c.UntilTimestamp.Time, err = time.Parse("2006-01-02", parts[1])
	case "tombstone":
		c.TombstoneType = parts[1]
	case "filter":
		switch parts[1] {
		case "links":
			c.FilterLinks = REQUIRE
		case "images":
			c.FilterImages = REQUIRE
		case "videos":
			c.FilterVideos = REQUIRE
		case "media":
			c.FilterMedia = REQUIRE
		case "polls":
			c.FilterPolls = REQUIRE
		case "spaces":
			c.FilterSpaces = REQUIRE
		case "replies":
			c.FilterReplies = REQUIRE
		case "retweets":
			c.FilterRetweets = REQUIRE
		}
	case "-filter":
		switch parts[1] {
		case "links":
			c.FilterLinks = EXCLUDE
		case "images":
			c.FilterImages = EXCLUDE
		case "videos":
			c.FilterVideos = EXCLUDE
		case "media":
			c.FilterMedia = EXCLUDE
		case "polls":
			c.FilterPolls = EXCLUDE
		case "spaces":
			c.FilterSpaces = EXCLUDE
		case "replies":
			c.FilterReplies = EXCLUDE
		case "retweets":
			c.FilterRetweets = EXCLUDE
		}
	}

	if err != nil {
		return fmt.Errorf("query token %q: %w", token, ErrInvalidQuery)
	}
	return nil
}

func (p Profile) NextPage(c Cursor, current_user_id UserID) (Feed, error) {
	where_clauses := []string{}
	bind_values := []interface{}{}

	// Keywords
	for _, kw := range c.Keywords {
		where_clauses = append(where_clauses, "text like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", kw))
	}

	// From, to, by, and RT'd by user handles
	if c.FromUserHandle != "" {
		where_clauses = append(where_clauses, "tweets.user_id = (select id from users_by_handle where handle like ?)")
		bind_values = append(bind_values, c.FromUserHandle)
	}
	for _, to_user := range c.ToUserHandles {
		where_clauses = append(where_clauses, "reply_mentions like ?")
		bind_values = append(bind_values, fmt.Sprintf("%%%s%%", to_user))
	}
	if c.RetweetedByUserHandle != "" {
		where_clauses = append(where_clauses, "retweeted_by = (select id from users_by_handle where handle like ?)")
		bind_values = append(bind_values, c.RetweetedByUserHandle)
	}
	if c.ByUserHandle != "" {
		where_clauses = append(where_clauses, "by_user_id = (select id from users_by_handle where handle like ?)")
		bind_values = append(bind_values, c.ByUserHandle)
	}
	if c.ListID != 0 {
		where_clauses = append(where_clauses, "by_user_id in (select user_id from list_users where list_id = ?)")
		bind_values = append(bind_values, c.ListID)
	}
	if c.FollowedByUserHandle != "" {
		where_clauses = append(where_clauses,
			"by_user_id in (select followee_id from follows where follower_id = (select id from users_by_handle where handle like ?))")
		bind_values = append(bind_values, c.FollowedByUserHandle)
	}
	if c.QuotedTweetID != 0 {
		where_clauses = append(where_clauses, "quoted_tweet_id = ?")
		bind_values = append(bind_values, c.QuotedTweetID)
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

	// Tombstone filter
	if c.TombstoneType == "true" {
		where_clauses = append(where_clauses, "tombstone_type != 0")
	} else if c.TombstoneType != "" {
		where_clauses = append(where_clauses, "tombstone_type = (select rowid from tombstone_types where short_name like ?)")
		bind_values = append(bind_values, c.TombstoneType)
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
	switch c.FilterMedia {
	case REQUIRE:
		where_clauses = append(where_clauses, `(exists (select 1 from videos where videos.tweet_id = tweets.id)
		                                     or exists (select 1 from images where images.tweet_id = tweets.id))`)
	case EXCLUDE:
		where_clauses = append(where_clauses, `not (exists (select 1 from videos where videos.tweet_id = tweets.id)
		                                         or exists (select 1 from images where images.tweet_id = tweets.id))`)
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

	liked_by_filter_join_clause := ""
	likes_sort_order_field := ""
	if c.LikedByUserHandle != "" {
		liked_by_filter_join_clause = " join likes filter_likes on tweets.id = filter_likes.tweet_id "
		where_clauses = append(where_clauses, "filter_likes.user_id = (select id from users_by_handle where handle like ?) ")
		bind_values = append(bind_values, c.LikedByUserHandle)
		likes_sort_order_field = ", coalesce(filter_likes.sort_order, -1) likes_sort_order "

		// Don't include retweets on "liked by" searches because it doesn't distinguish which retweet
		// version was the "liked" one
		where_clauses = append(where_clauses, "retweet_id = 0")
	}

	bookmarked_by_filter_join_clause := ""
	bookmarks_sort_order_field := ""
	if c.BookmarkedByUserHandle != "" {
		bookmarked_by_filter_join_clause = " join bookmarks filter_bookmarks on tweets.id = filter_bookmarks.tweet_id "
		where_clauses = append(where_clauses, "filter_bookmarks.user_id = (select id from users_by_handle where handle like ?) ")
		bind_values = append(bind_values, c.BookmarkedByUserHandle)
		bookmarks_sort_order_field = ", coalesce(filter_bookmarks.sort_order, -1) bookmarks_sort_order "

		// Don't include retweets on "bookmarked by" searches because it doesn't distinguish which retweet
		// version was the "bookmarked" one
		where_clauses = append(where_clauses, "retweet_id = 0")
	}

	// Pagination
	if c.CursorPosition != CURSOR_START {
		where_clauses = append(where_clauses, c.SortOrder.PaginationWhereClause())
		bind_values = append(bind_values, c.CursorValue)
	}

	// Assemble the full where-clause
	where_clause := ""
	if len(where_clauses) > 0 {
		where_clause = "where " + strings.Join(where_clauses, " and ")
	}

	// The Query:
	//   1. Base query:
	//     a. Include "likes_sort_order" and "bookmarks_sort_order" fields, if they're in the filters
	//     b. Left join on "likes" table to get whether logged-in user has liked the tweet
	//     c. Left join on "likes" and "bookmarks" tables, if needed (i.e., if in the filters)
	//     d. Add 'where', 'order by', and (mildly unnecessary) 'limit' clauses
	//   2. Two copies of the base query, one for "tweets" and one for "retweets", joined with "union"
	//   3. Actual "limit" clause
	q := `select * from (
	select ` + TWEETS_ALL_SQL_FIELDS + `,
	       exists (select 1 from retweets where tweet_id = tweets.id and retweeted_by = ?) is_retweeted_by_current_user` +
		likes_sort_order_field + bookmarks_sort_order_field + `,
           0 tweet_id, 0 retweet_id, 0 retweeted_by, 0 retweeted_at,
           posted_at chrono, tweets.user_id by_user_id
      from tweets
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
 left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
     ` + liked_by_filter_join_clause + `
     ` + bookmarked_by_filter_join_clause + `
     ` + where_clause + ` ` + c.SortOrder.OrderByClause() + ` limit ?
    )

     union

    select * from (
    select ` + TWEETS_ALL_SQL_FIELDS + `,
           exists (select 1 from retweets where tweet_id = tweets.id and retweeted_by = ?) is_retweeted_by_current_user` +
		likes_sort_order_field + bookmarks_sort_order_field + `,
           retweets.tweet_id, retweet_id, retweeted_by, retweeted_at,
           retweeted_at chrono, retweeted_by by_user_id
      from retweets
 left join tweets on retweets.tweet_id = tweets.id
 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
 left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
     ` + liked_by_filter_join_clause + `
     ` + bookmarked_by_filter_join_clause + `
     ` + where_clause + ` ` + c.SortOrder.OrderByClause() + ` limit ?
   )
   ` + c.SortOrder.OrderByClause() + ` limit ?`

	bind_values = append([]interface{}{current_user_id, current_user_id}, bind_values...)
	bind_values = append(bind_values, c.PageSize)
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
		// fmt.Printf("\tResult: %#v\n", val)
		ret.Tweets[val.Tweet.ID] = val.Tweet
		if val.Retweet.RetweetID != 0 {
			ret.Retweets[val.Retweet.RetweetID] = val.Retweet
		}
		ret.Items = append(ret.Items, FeedItem{TweetID: val.Tweet.ID, RetweetID: val.Retweet.RetweetID})
	}

	p.fill_content(&ret.TweetTrove, current_user_id)

	// Set the new cursor position and value
	ret.CursorBottom = c // Copy cursor values over
	if len(results) < c.PageSize {
		ret.CursorBottom.CursorPosition = CURSOR_END
	} else {
		ret.CursorBottom.CursorPosition = CURSOR_MIDDLE
		last_item := results[len(results)-1]
		ret.CursorBottom.CursorValue = c.SortOrder.NextCursorValue(last_item)
	}

	return ret, nil
}
