package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

var (
	ErrEndOfFeed = errors.New("end of feed")
	ErrNotInDB   = errors.New("not in database")
)

func (p Profile) fill_content(trove *TweetTrove) {
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
		     select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
		            mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id,
		            ifnull(tombstone_types.short_name, "") tombstone_type, is_expandable,
		            is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at
		       from tweets
		  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
		      where id in (`+strings.Repeat("?,", len(quoted_ids)-1)+`?)`, quoted_ids...)
		if err != nil {
			panic(err)
		}
		for _, t := range quoted_tweets {
			trove.Tweets[t.ID] = t
		}
	}

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

	// Get all the users
	if len(user_ids) > 0 { // It could be a search with no results, end of feed, etc-- strings.Repeat will fail!
		var users []User
		userquery := `
	        select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
	               is_banned, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path, pinned_tweet_id,
	               is_content_downloaded, is_followed
	          from users
	         where id in (` + strings.Repeat("?,", len(user_ids)-1) + `?)`
		fmt.Printf("%s\n", userquery)
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
	fmt.Printf("%s\n", imgquery) // TODO: SQL logger
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
func (p Profile) GetTweetDetail(id TweetID) (TweetDetailView, error) {
	// TODO: compound-query-structs
	ret := NewTweetDetailView()

	stmt, err := p.DB.Preparex(
		`select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
	            mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
	            is_expandable,
	            is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at
	       from tweets
	  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
	      where id = ?`)
	if err != nil {
		panic(err)
	}

	// Main tweet
	var tweet Tweet
	err = stmt.Get(&tweet, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ret, fmt.Errorf("Tweet ID %d: %w", id, ErrNotInDB)
		} else {
			panic(err)
		}
	}
	ret.Tweets[tweet.ID] = tweet
	ret.MainTweetID = tweet.ID

	// Parent tweets
	for tweet.InReplyToID != 0 {
		err := stmt.Get(&tweet, tweet.InReplyToID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			} else {
				panic(err)
			}
		}
		ret.Tweets[tweet.ID] = tweet
		ret.ParentIDs = append([]TweetID{tweet.ID}, ret.ParentIDs...)
	}

	// TODO: tweet-detail-thread-chains

	// Replies lvl 1
	// TODO: this can be done with a recursive Common Table Expression:
	// - https://www.sqlite.org/lang_with.html#recursive_query_examples
	var replies []Tweet
	stmt, err = p.DB.Preparex(
		`select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
	            mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
	            is_expandable,
	            is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at
	       from tweets
	  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
	      where in_reply_to_id = ?
	      order by num_likes desc
	      limit 50`)
	if err != nil {
		panic(err)
	}
	err = stmt.Select(&replies, id)
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

		    select tweets.id id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id,
		           quoted_tweet_id, mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id,
		           ifnull(tombstone_types.short_name, "") tombstone_type, is_expandable,
		           is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at
		      from top_ids_by_parent
		 left join tweets on tweets.id = top_ids_by_parent.id
		 left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid`
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

	p.fill_content(&ret.TweetTrove)
	return ret, nil
}

// TODO: compound-query-structs
type FeedItem struct {
	TweetID
	RetweetID TweetID
}
type Feed struct {
	Items []FeedItem
	TweetTrove
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

// Return the given tweet, all its parent tweets, and a list of conversation threads
func (p Profile) GetUserFeed(id UserID, count int, max_posted_at Timestamp) (Feed, error) {
	ret := NewFeed()

	tweet_max_clause := ""
	retweet_max_clause := ""
	if max_posted_at.Unix() > 0 {
		tweet_max_clause = " and posted_at < :max_posted_at "
		retweet_max_clause = " and retweeted_at < :max_posted_at "
	}

	q := `select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
               mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
               is_expandable,
               is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
               0 tweet_id, 0 retweet_id, 0 retweeted_by, 0 retweeted_at,
               posted_at order_by
          from tweets
     left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
         where user_id = :id` + tweet_max_clause + `

         union

        select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
               mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
               is_expandable,
               is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at,
               tweet_id, retweet_id, retweeted_by, retweeted_at,
               retweeted_at order_by
          from retweets
     left join tweets on retweets.tweet_id = tweets.id
     left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
         where retweeted_by = :id` + retweet_max_clause + `

         order by order_by desc
         limit :limit`

	stmt, err := p.DB.PrepareNamed(q)
	if err != nil {
		panic(err)
	}

	args := map[string]interface{}{
		"id":            id,
		"limit":         count,
		"max_posted_at": max_posted_at,
	}
	var results []struct {
		Tweet
		Retweet
		OrderBy int `db:"order_by"`
	}
	err = stmt.Select(&results, args)
	if err != nil {
		panic(err)
	}
	if len(results) == 0 {
		return NewFeed(), ErrEndOfFeed
	}

	for _, val := range results {
		ret.Tweets[val.Tweet.ID] = val.Tweet
		if val.Retweet.RetweetID != 0 {
			ret.Retweets[val.Retweet.RetweetID] = val.Retweet
		}
		ret.Items = append(ret.Items, FeedItem{TweetID: val.Tweet.ID, RetweetID: val.Retweet.RetweetID})
	}

	p.fill_content(&ret.TweetTrove)

	return ret, nil
}
