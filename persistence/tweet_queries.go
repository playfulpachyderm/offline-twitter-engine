package persistence

import (
	"database/sql"
	"strings"

	"offline_twitter/scraper"
)

func (p Profile) SaveTweet(t scraper.Tweet) error {
	db := p.DB

	tx := db.MustBegin()

	_, err := db.Exec(`
        insert into tweets (id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id,
                            quoted_tweet_id, mentions, reply_mentions, hashtags, tombstone_type, is_stub, is_content_downloaded,
                            is_conversation_scraped, last_scraped_at)
        values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, (select rowid from tombstone_types where short_name=?), ?, ?, ?, ?)
            on conflict do update
           set num_likes=?,
               num_retweets=?,
               num_replies=?,
               num_quote_tweets=?,
               is_stub=(is_stub and ?),
               is_content_downloaded=(is_content_downloaded or ?),
               is_conversation_scraped=(is_conversation_scraped or ?),
               last_scraped_at=max(last_scraped_at, ?)
        `,
		t.ID, t.UserID, t.Text, t.PostedAt, t.NumLikes, t.NumRetweets, t.NumReplies, t.NumQuoteTweets, t.InReplyToID,
		t.QuotedTweetID, scraper.JoinArrayOfHandles(t.Mentions), scraper.JoinArrayOfHandles(t.ReplyMentions),
		strings.Join(t.Hashtags, ","), t.TombstoneType, t.IsStub, t.IsContentDownloaded, t.IsConversationScraped, t.LastScrapedAt,

		t.NumLikes, t.NumRetweets, t.NumReplies, t.NumQuoteTweets, t.IsStub, t.IsContentDownloaded, t.IsConversationScraped,
		t.LastScrapedAt,
	)

	if err != nil {
		return err
	}
	for _, url := range t.Urls {
		err := p.SaveUrl(url)
		if err != nil {
			return err
		}
	}
	for _, image := range t.Images {
		err := p.SaveImage(image)
		if err != nil {
			return err
		}
	}
	for _, video := range t.Videos {
		err := p.SaveVideo(video)
		if err != nil {
			return err
		}
	}
	for _, hashtag := range t.Hashtags {
		_, err := db.Exec("insert into hashtags (tweet_id, text) values (?, ?) on conflict do nothing", t.ID, hashtag)
		if err != nil {
			return err
		}
	}
	for _, poll := range t.Polls {
		err := p.SavePoll(poll)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p Profile) IsTweetInDatabase(id scraper.TweetID) bool {
	db := p.DB

	var dummy string
	err := db.QueryRow("select 1 from tweets where id = ?", id).Scan(&dummy)
	if err != nil {
		if err != sql.ErrNoRows {
			// A real error
			panic(err)
		}
		return false
	}
	return true
}

func (p Profile) GetTweetById(id scraper.TweetID) (scraper.Tweet, error) {
	db := p.DB

	stmt, err := db.Prepare(`
        select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
               mentions, reply_mentions, hashtags, ifnull(tombstone_types.short_name, ""), is_stub, is_content_downloaded,
               is_conversation_scraped, last_scraped_at
          from tweets left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
         where id = ?
    `)

	if err != nil {
		return scraper.Tweet{}, err
	}
	defer stmt.Close()

	var t scraper.Tweet
	var mentions string
	var reply_mentions string
	var hashtags string

	row := stmt.QueryRow(id)
	err = row.Scan(&t.ID, &t.UserID, &t.Text, &t.PostedAt, &t.NumLikes, &t.NumRetweets, &t.NumReplies, &t.NumQuoteTweets, &t.InReplyToID,
		&t.QuotedTweetID, &mentions, &reply_mentions, &hashtags, &t.TombstoneType, &t.IsStub, &t.IsContentDownloaded,
		&t.IsConversationScraped, &t.LastScrapedAt)
	if err != nil {
		return t, err
	}

	t.Mentions = []scraper.UserHandle{}
	for _, m := range strings.Split(mentions, ",") {
		if m != "" {
			t.Mentions = append(t.Mentions, scraper.UserHandle(m))
		}
	}
	t.ReplyMentions = []scraper.UserHandle{}
	for _, m := range strings.Split(reply_mentions, ",") {
		if m != "" {
			t.ReplyMentions = append(t.ReplyMentions, scraper.UserHandle(m))
		}
	}
	t.Hashtags = []string{}
	for _, h := range strings.Split(hashtags, ",") {
		if h != "" {
			t.Hashtags = append(t.Hashtags, h)
		}
	}

	imgs, err := p.GetImagesForTweet(t)
	if err != nil {
		return t, err
	}
	t.Images = imgs

	vids, err := p.GetVideosForTweet(t)
	if err != nil {
		return t, err
	}
	t.Videos = vids

	polls, err := p.GetPollsForTweet(t)
	if err != nil {
		return t, err
	}
	t.Polls = polls

	urls, err := p.GetUrlsForTweet(t)
	t.Urls = urls

	return t, err
}

/**
 * Populate the `User` field on a tweet with an actual User
 */
func (p Profile) LoadUserFor(t *scraper.Tweet) error {
	if t.User != nil {
		// Already there, no need to load it
		return nil
	}

	user, err := p.GetUserByID(t.UserID)
	if err != nil {
		return err
	}
	t.User = &user
	return nil
}

/**
 * Return `false` if the tweet is in the DB and has had its content downloaded, `false` otherwise
 */
func (p Profile) CheckTweetContentDownloadNeeded(tweet scraper.Tweet) bool {
	row := p.DB.QueryRow(`select is_content_downloaded from tweets where id = ?`, tweet.ID)

	var is_content_downloaded bool
	err := row.Scan(&is_content_downloaded)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		} else {
			panic(err)
		}
	}
	return !is_content_downloaded
}
