package persistence

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

var (
	ErrEndOfFeed = errors.New("end of feed")
)

// TODO: make this a SQL view?
const TWEETS_ALL_SQL_FIELDS = `
		tweets.id id, tweets.user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id,
		quoted_tweet_id, mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id,
		ifnull(tombstone_types.short_name, "") tombstone_type, ifnull(tombstone_types.tombstone_text, "") tombstone_text,
		case when likes.user_id is null then 0 else 1 end is_liked_by_current_user,
		is_expandable, is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at`

// Given a TweetTrove, fetch its:
// - quoted tweets
// - spaces
// - users
// - images, videos, urls, polls
func (p Profile) fill_content(trove *TweetTrove, current_user_id UserID) {
	if len(trove.Tweets) == 0 {
		// Empty trove, nothing to fetch
		return
	}

	// Fetch quote-tweets
	// TODO: use recursive Common Table Expressions?
	quoted_ids := []interface{}{}
	for _, t := range trove.Tweets {
		if t.QuotedTweetID != 0 {
			quoted_ids = append(quoted_ids, t.QuotedTweetID)
		}
	}
	if len(quoted_ids) > 0 {
		var quoted_tweets []Tweet
		err := p.DB.Select(&quoted_tweets, `
		     select `+TWEETS_ALL_SQL_FIELDS+`
		       from tweets
		  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
		  left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
		      where id in (`+strings.Repeat("?,", len(quoted_ids)-1)+`?)`, append([]interface{}{current_user_id}, quoted_ids...)...)
		if err != nil {
			panic(err)
		}
		for _, t := range quoted_tweets {
			trove.Tweets[t.ID] = t
		}
	}

	// Fetch spaces
	space_ids := []interface{}{}
	for _, t := range trove.Tweets {
		if t.SpaceID != "" {
			space_ids = append(space_ids, t.SpaceID)
		}
	}
	if len(space_ids) > 0 {
		var spaces []Space
		err := p.DB.Select(&spaces, `
		 select id, ifnull(created_by_id, 0) created_by_id, short_url, state, title, ifnull(created_at, 0) created_at,
		        ifnull(started_at, 0) started_at, ifnull(ended_at, 0) ended_at, ifnull(updated_at, 0) updated_at,
		        ifnull(is_available_for_replay, 0) is_available_for_replay, ifnull(replay_watch_count, 0) replay_watch_count,
		        ifnull(live_listeners_count, 0) replay_watch_count, is_details_fetched
		   from spaces
		  where id in (`+strings.Repeat("?,", len(space_ids)-1)+`?)`,
			space_ids...,
		)
		if err != nil {
			panic(err)
		}
		for _, s := range spaces {
			err := p.DB.Select(&s.ParticipantIds, "select user_id from space_participants where space_id = ?", s.ID)
			if err != nil {
				panic(err)
			}
			trove.Spaces[s.ID] = s
		}
	}

	// Assemble list of users fetched in previous operations
	in_clause := ""
	user_ids := []interface{}{}
	tweet_ids := []interface{}{}
	for _, t := range trove.Tweets {
		in_clause += "?,"
		user_ids = append(user_ids, int(t.UserID))
		tweet_ids = append(tweet_ids, t.ID)
	}
	in_clause = in_clause[:len(in_clause)-1]

	for _, r := range trove.Retweets {
		user_ids = append(user_ids, int(r.RetweetedByID))
	}
	for _, s := range trove.Spaces {
		user_ids = append(user_ids, s.CreatedById)
		for _, p := range s.ParticipantIds {
			user_ids = append(user_ids, p)
		}
	}
	for _, n := range trove.Notifications {
		// Primary user
		if n.ActionUserID != UserID(0) {
			user_ids = append(user_ids, n.ActionUserID)
		}
		// Other users, if there are any
		for _, u_id := range n.UserIDs {
			user_ids = append(user_ids, u_id)
		}
	}

	// Get all the users
	if len(user_ids) > 0 { // It could be a search with no results, end of feed, etc-- strings.Repeat will fail!
		var users []User
		userquery := `
	        select ` + USERS_ALL_SQL_FIELDS + `
	          from users
	         where id in (` + strings.Repeat("?,", len(user_ids)-1) + `?)`
		// fmt.Printf("%s\n", userquery)
		err := p.DB.Select(&users, userquery, user_ids...)
		if err != nil {
			panic(err)
		}
		for _, u := range users {
			trove.Users[u.ID] = u
		}
	}

	// Get all the Images
	var images []Image
	imgquery := `
        select id, tweet_id, width, height, remote_url, local_filename, is_downloaded from images where tweet_id in (` + in_clause + `)`
	err := p.DB.Select(&images, imgquery, tweet_ids...)
	if err != nil {
		panic(err)
	}
	for _, i := range images {
		t, is_ok := trove.Tweets[i.TweetID]
		if !is_ok {
			panic(i)
		}
		t.Images = append(t.Images, i)
		trove.Tweets[t.ID] = t
	}

	// Get all the Videos
	var videos []Video
	err = p.DB.Select(&videos, `
        select id, tweet_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename, duration,
		       view_count, is_downloaded, is_blocked_by_dmca, is_gif
		  from videos
		 where tweet_id in (`+in_clause+`)`, tweet_ids...)
	if err != nil {
		panic(err)
	}
	for _, v := range videos {
		t, is_ok := trove.Tweets[v.TweetID]
		if !is_ok {
			panic(v)
		}
		t.Videos = append(t.Videos, v)
		trove.Tweets[t.ID] = t
	}

	// Get all the Urls
	var urls []Url
	err = p.DB.Select(&urls, `
        select tweet_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
		       thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded
		  from urls
		 where tweet_id in (`+in_clause+`)`, tweet_ids...)
	if err != nil {
		panic(err)
	}
	for _, u := range urls {
		t, is_ok := trove.Tweets[u.TweetID]
		if !is_ok {
			panic(u)
		}
		t.Urls = append(t.Urls, u)
		trove.Tweets[t.ID] = t
	}

	// Get all the Polls
	var polls []Poll
	err = p.DB.Select(&polls, `
		select id, tweet_id, num_choices, choice1, choice1_votes, choice2, choice2_votes, choice3, choice3_votes, choice4, choice4_votes,
		       voting_duration, voting_ends_at, last_scraped_at
		  from polls
		 where tweet_id in (`+in_clause+`)`, tweet_ids...)
	if err != nil {
		panic(err)
	}
	for _, p := range polls {
		t, is_ok := trove.Tweets[p.TweetID]
		if !is_ok {
			panic(p)
		}
		t.Polls = append(t.Polls, p)
		trove.Tweets[t.ID] = t
	}
}

// TODO: compound-query-structs
type TweetDetailView struct {
	TweetTrove
	ParentIDs   []TweetID
	MainTweetID TweetID
	ThreadIDs   []TweetID
	ReplyChains [][]TweetID
}

func NewTweetDetailView() TweetDetailView {
	return TweetDetailView{
		TweetTrove:  NewTweetTrove(),
		ParentIDs:   []TweetID{},
		ReplyChains: [][]TweetID{},
	}
}

// Return the given tweet, all its parent tweets, and a list of conversation threads
func (p Profile) GetTweetDetail(id TweetID, current_user_id UserID) (TweetDetailView, error) {
	// TODO: compound-query-structs
	ret := NewTweetDetailView()
	ret.MainTweetID = id

	stmt, err := p.DB.Preparex(`
		   with recursive all_replies(id) as (values(?) union all
		        select tweets.in_reply_to_id from tweets, all_replies
		         where tweets.id = all_replies.id and tweets.in_reply_to_id != 0
		        )

		 select ` + TWEETS_ALL_SQL_FIELDS + `
	       from tweets
	  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
	  left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
	 inner join all_replies on tweets.id = all_replies.id
	      order by id asc`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Main tweet and parents
	var thread []Tweet
	err = stmt.Select(&thread, id, current_user_id)
	if err != nil {
		panic(err)
	}
	if len(thread) == 0 {
		return ret, fmt.Errorf("Tweet ID %d: %w", id, ErrNotInDatabase)
	}
	for _, tweet := range thread {
		ret.Tweets[tweet.ID] = tweet
		if tweet.ID != ret.MainTweetID {
			ret.ParentIDs = append(ret.ParentIDs, tweet.ID)
		}
	}

	// Threaded replies
	stmt, err = p.DB.Preparex(`
			with recursive thread_replies(id) as (
				values(?)
					union all
				select tweets.id from tweets
				                 join thread_replies on tweets.in_reply_to_id = thread_replies.id
				                where tweets.user_id = ?
			)

		 select ` + TWEETS_ALL_SQL_FIELDS + `
	       from tweets
	  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
	  left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
	 inner join thread_replies on tweets.id = thread_replies.id
	      order by id asc`)

	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	var reply_thread []Tweet
	err = stmt.Select(&reply_thread, id, ret.Tweets[ret.MainTweetID].UserID, current_user_id)
	if err != nil {
		panic(err)
	}
	for _, tweet := range reply_thread {
		ret.Tweets[tweet.ID] = tweet
		if tweet.ID != ret.MainTweetID {
			ret.ThreadIDs = append(ret.ThreadIDs, tweet.ID)
		}
	}

	var replies []Tweet
	stmt, err = p.DB.Preparex(
		`select ` + TWEETS_ALL_SQL_FIELDS + `
	       from tweets
	  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
	  left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
	      where in_reply_to_id = ?
	        and id != ? -- skip the main Thread if there is one
	      order by num_likes desc
	      limit 50`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	thread_top_id := TweetID(0)
	if len(ret.ThreadIDs) > 0 {
		thread_top_id = ret.ThreadIDs[0]
	}
	err = stmt.Select(&replies, current_user_id, id, thread_top_id)
	if err != nil {
		panic(err)
	}

	if len(replies) > 0 {
		reply_1_ids := []interface{}{}
		for _, r := range replies {
			ret.Tweets[r.ID] = r
			reply_1_ids = append(reply_1_ids, r.ID)
			ret.ReplyChains = append(ret.ReplyChains, []TweetID{r.ID})
		}
		reply2_query := `
		      with parent_ids(id) as (values ` + strings.Repeat("(?), ", len(reply_1_ids)-1) + `(?)),
		           all_reply_ids(id, parent_id, num_likes) as (
		               select tweets.id, tweets.in_reply_to_id, num_likes
		                 from parent_ids
		                 left join tweets on tweets.in_reply_to_id = parent_ids.id
		           ),
		           top_ids_by_parent(id, parent_id) as (
		               select id, parent_id outer_parent_id
		                 from all_reply_ids
		                where id = (
		                    select id from all_reply_ids
		                     where parent_id = outer_parent_id
		                  order by num_likes desc limit 1
		                )
		           )

		    select ` + TWEETS_ALL_SQL_FIELDS + `
		      from top_ids_by_parent
		 left join tweets on tweets.id = top_ids_by_parent.id
		 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
		 left join likes on tweets.id = likes.tweet_id and likes.user_id = ?`
		reply_1_ids = append(reply_1_ids, current_user_id)
		err = p.DB.Select(&replies, reply2_query, reply_1_ids...)
		if err != nil {
			panic(err)
		}
		for _, r := range replies {
			ret.Tweets[r.ID] = r
			for i, chain := range ret.ReplyChains {
				if len(chain) == 1 && chain[0] == r.InReplyToID {
					ret.ReplyChains[i] = append(chain, r.ID)
					break
				}
				// TODO: Log weird situation
			}
		}
	}

	p.fill_content(&ret.TweetTrove, current_user_id)
	return ret, nil
}

// TODO: compound-query-structs
type FeedItem struct {
	TweetID
	RetweetID TweetID
	NotificationID
	QuoteNestingLevel int // Defines the current nesting level (not available remaining levels)
}
type Feed struct {
	Items []FeedItem
	TweetTrove
	CursorBottom Cursor
}

func (f Feed) BottomTimestamp() Timestamp {
	if len(f.Items) == 0 {
		return TimestampFromUnix(0)
	}
	last := f.Items[len(f.Items)-1]
	if last.RetweetID != 0 {
		return f.Retweets[last.RetweetID].RetweetedAt
	}
	return f.Tweets[last.TweetID].PostedAt
}

func NewFeed() Feed {
	return Feed{
		Items:      []FeedItem{},
		TweetTrove: NewTweetTrove(),
	}
}

func (p Profile) GetNotificationsForUser(u_id UserID, cursor int64, count int64) Feed {
	// Get the notifications
	var notifications []Notification
	err := p.DB.Select(&notifications,
		`select id, type, sent_at, sort_index, user_id, ifnull(action_user_id, 0) action_user_id,
		        ifnull(action_tweet_id, 0) action_tweet_id, ifnull(action_retweet_id, 0) action_retweet_id, has_detail, last_scraped_at
		   from notifications
		  where (sort_index < ? or ?)
		    and user_id = ?
		  order by sort_index desc
		  limit ?
	`, cursor, cursor == 0, u_id, count)
	if err != nil {
		panic(err)
	}

	// Get the user_ids list for each notification.  Unlike tweet+retweet_ids, users are needed to render
	// the notification properly.
	for i := range notifications {
		err = p.DB.Select(&notifications[i].UserIDs,
			`select user_id from notification_users where notification_id = ?`,
			notifications[i].ID,
		)
		if err != nil {
			panic(err)
		}
	}

	// Collect tweet and retweet IDs
	retweet_ids := []TweetID{}
	tweet_ids := []TweetID{}
	for _, n := range notifications {
		if n.ActionRetweetID != TweetID(0) {
			retweet_ids = append(retweet_ids, n.ActionRetweetID)
		}
		if n.ActionTweetID != TweetID(0) {
			tweet_ids = append(tweet_ids, n.ActionTweetID)
		}
	}

	// TODO: can this go in `fill_content`?

	// Get retweets if there are any
	var retweets []Retweet
	if len(retweet_ids) != 0 {
		sql_str, vals, err := sqlx.In(`
			select retweet_id, tweet_id, retweeted_by, retweeted_at
			  from retweets
			 where retweet_id in (?)
		`, retweet_ids)
		if err != nil {
			panic(err)
		}
		err = p.DB.Select(&retweets, sql_str, vals...)
		if err != nil {
			panic(err)
		}

		// Collect more tweet IDs, from retweets
		for _, r := range retweets {
			tweet_ids = append(tweet_ids, r.TweetID)
		}
	}

	// Get tweets, if there are any
	var tweets []Tweet
	if len(tweet_ids) != 0 {
		sql_str, vals, err := sqlx.In(`select `+TWEETS_ALL_SQL_FIELDS+`
			from tweets
			left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
			left join likes on tweets.id = likes.tweet_id and likes.user_id = -1
          where id in (?)`, tweet_ids)
		if err != nil {
			panic(err)
		}
		err = p.DB.Select(&tweets, sql_str, vals...)
		if err != nil {
			panic(err)
		}
	}

	ret := NewFeed()
	for _, t := range tweets {
		ret.TweetTrove.Tweets[t.ID] = t
	}
	for _, r := range retweets {
		ret.TweetTrove.Retweets[r.RetweetID] = r
	}
	for _, n := range notifications {
		// Add to tweet trove
		ret.TweetTrove.Notifications[n.ID] = n

		// Construct feed item
		feed_item := FeedItem{
			NotificationID: n.ID,
			RetweetID:      n.ActionRetweetID, // might be 0
			TweetID:        n.ActionTweetID,   // might be 0
		}
		r, is_ok := ret.TweetTrove.Retweets[n.ActionRetweetID]
		if is_ok {
			// If the action has a retweet, fill the FeedItem.TweetID from the retweet
			feed_item.TweetID = r.TweetID
		}
		ret.Items = append(ret.Items, feed_item)
	}

	// TODO: proper user id
	p.fill_content(&ret.TweetTrove, UserID(0))

	// Set the bottom cursor value
	ret.CursorBottom = Cursor{}
	if len(ret.Items) < int(count) {
		ret.CursorBottom.CursorPosition = CURSOR_END
	} else {
		ret.CursorBottom.CursorPosition = CURSOR_MIDDLE
		last_item := ret.Items[len(ret.Items)-1]
		last_notif, is_ok := ret.Notifications[last_item.NotificationID]
		if !is_ok {
			panic("last item isn't a notification???")
		}
		ret.CursorBottom.CursorValue = int(last_notif.SortIndex) // TODO: CursorValue should be int64
	}
	return ret
}
