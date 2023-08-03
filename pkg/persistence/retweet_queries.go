package persistence

import (
	"fmt"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Save a Retweet.  Do nothing if it already exists, because none of its parameters are modifiable.
func (p Profile) SaveRetweet(r scraper.Retweet) error {
	_, err := p.DB.NamedExec(`
			insert into retweets (retweet_id, tweet_id, retweeted_by, retweeted_at)
			values (:retweet_id, :tweet_id, :retweeted_by, :retweeted_at)
			    on conflict do nothing
		`,
		r,
	)
	if err != nil {
		return fmt.Errorf("Error executing SaveRetweet(%#v):\n  %w", r, err)
	}
	return nil
}

// Retrieve a Retweet by ID
func (p Profile) GetRetweetById(id scraper.TweetID) (scraper.Retweet, error) {
	var r scraper.Retweet
	err := p.DB.Get(&r, `
		select retweet_id, tweet_id, retweeted_by, retweeted_at
		  from retweets
		 where retweet_id = ?
	`, id)
	if err != nil {
		return r, fmt.Errorf("Error executing GetRetweetById(%d):\n  %w", id, err)
	}
	return r, nil
}
