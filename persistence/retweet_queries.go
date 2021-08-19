package persistence

import (
	"time"

	"offline_twitter/scraper"
)

/**
 * Save a Retweet.  Do nothing if it already exists, because none of its parameters are modifiable.
 */
func (p Profile) SaveRetweet(r scraper.Retweet) error {
	_, err := p.DB.Exec(`
			insert into retweets (retweet_id, tweet_id, retweeted_by, retweeted_at)
			values (?, ?, ?, ?)
			    on conflict do nothing
		`,
		r.RetweetID, r.TweetID, r.RetweetedByID, r.RetweetedAt.Unix(),
	)
	return err
}


/**
 * Retrieve a Retweet by ID
 */
func (p Profile) GetRetweetById(id scraper.TweetID) (scraper.Retweet, error) {
	stmt, err := p.DB.Prepare(`
		select retweet_id, tweet_id, retweeted_by, retweeted_at
		  from retweets
		 where retweet_id = ?
	`)
	if err != nil {
		return scraper.Retweet{}, err
	}
	defer stmt.Close()

	var r scraper.Retweet
	var retweeted_at int

	row := stmt.QueryRow(id)
	err = row.Scan(&r.RetweetID, &r.TweetID, &r.RetweetedByID, &retweeted_at)
	if err != nil {
		return scraper.Retweet{}, err
	}

	r.RetweetedAt = time.Unix(int64(retweeted_at), 0)
	return r, nil
}