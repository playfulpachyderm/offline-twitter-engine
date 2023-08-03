package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)


func (p Profile) SaveTweet(t scraper.Tweet) error {
	db := p.DB

	tx := db.MustBegin()

	// Has to be done first since Tweet has a foreign key to Space
	for _, space := range t.Spaces {
		err := p.SaveSpace(space)
		if err != nil {
			return err
		}
	}

	_, err := db.NamedExec(`
        insert into tweets (id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id,
                            quoted_tweet_id, mentions, reply_mentions, hashtags, space_id, tombstone_type, is_expandable,
                            is_stub, is_content_downloaded,
                            is_conversation_scraped, last_scraped_at)
        values (:id, :user_id, :text, :posted_at, :num_likes, :num_retweets, :num_replies, :num_quote_tweets, :in_reply_to_id,
                :quoted_tweet_id, :mentions, :reply_mentions, :hashtags, nullif(:space_id, ''),
                (select rowid from tombstone_types where short_name=:tombstone_type),
                :is_expandable,
                :is_stub, :is_content_downloaded,
                :is_conversation_scraped, :last_scraped_at)
            on conflict do update
           set text=(case
                     when is_stub then
                         :text
                     when not is_expandable and :is_expandable then
                         :text
                     else
                         text
                     end
               ),
               num_likes=(case when :is_stub then num_likes else :num_likes end),
               num_retweets=(case when :is_stub then num_retweets else :num_retweets end),
               num_replies=(case when :is_stub then num_replies else :num_replies end),
               num_quote_tweets=(case when :is_stub then num_quote_tweets else :num_quote_tweets end),
               is_stub=(is_stub and :is_stub),
               tombstone_type=(case
                               when :tombstone_type='unavailable' and tombstone_type not in (0, 4) then
                                   tombstone_type
                               else
                                   (select rowid from tombstone_types where short_name=:tombstone_type)
                               end
               ),
               is_expandable=is_expandable or :is_expandable,
               is_content_downloaded=(is_content_downloaded or :is_content_downloaded),
               is_conversation_scraped=(is_conversation_scraped or :is_conversation_scraped),
               last_scraped_at=max(last_scraped_at, :last_scraped_at)
        `,
		t,
	)

	if err != nil {
		return fmt.Errorf("Error executing SaveTweet(ID %d).  Info: %#v:\n  %w", t.ID, t, err)
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
			return fmt.Errorf("Error inserting hashtag %q on tweet ID %d:\n  %w", hashtag, t.ID, err)
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
		return fmt.Errorf("Error committing SaveTweet transaction:\n  %w", err)
	}
	return nil
}

func (p Profile) IsTweetInDatabase(id scraper.TweetID) bool {
	db := p.DB

	var dummy string
	err := db.QueryRow("select 1 from tweets where id = ?", id).Scan(&dummy)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// A real error
			panic(err)
		}
		return false
	}
	return true
}

func (p Profile) GetTweetById(id scraper.TweetID) (scraper.Tweet, error) {
	db := p.DB

	var t scraper.Tweet
	err := db.Get(&t, `
        select id, user_id, text, posted_at, num_likes, num_retweets, num_replies, num_quote_tweets, in_reply_to_id, quoted_tweet_id,
               mentions, reply_mentions, hashtags, ifnull(space_id, '') space_id, ifnull(tombstone_types.short_name, "") tombstone_type,
               is_expandable,
               is_stub, is_content_downloaded, is_conversation_scraped, last_scraped_at
          from tweets left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
         where id = ?
    `, id)

	if err != nil {
		return scraper.Tweet{}, fmt.Errorf("Error executing GetTweetByID(%d):\n  %w", id, err)
	}

	t.Spaces = []scraper.Space{}
	if t.SpaceID != "" {
		space, err := p.GetSpaceById(t.SpaceID)
		if err != nil {
			return t, fmt.Errorf("Error retrieving space with ID %s (tweet %d):\n  %w", t.SpaceID, t.ID, err)
		}
		t.Spaces = append(t.Spaces, space)
	}

	imgs, err := p.GetImagesForTweet(t)
	if err != nil {
		return t, fmt.Errorf("Error retrieving images for tweet %d:\n  %w", t.ID, err)
	}
	t.Images = imgs

	vids, err := p.GetVideosForTweet(t)
	if err != nil {
		return t, fmt.Errorf("Error retrieving videos for tweet %d:\n  %w", t.ID, err)
	}
	t.Videos = vids

	polls, err := p.GetPollsForTweet(t)
	if err != nil {
		return t, fmt.Errorf("Error retrieving polls for tweet %d:\n  %w", t.ID, err)
	}
	t.Polls = polls

	urls, err := p.GetUrlsForTweet(t)
	if err != nil {
		return t, fmt.Errorf("Error retrieving urls for tweet %d:\n  %w", t.ID, err)
	}
	t.Urls = urls

	return t, nil
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
		if errors.Is(err, sql.ErrNoRows) {
			return true
		} else {
			panic(err)
		}
	}
	return !is_content_downloaded
}
