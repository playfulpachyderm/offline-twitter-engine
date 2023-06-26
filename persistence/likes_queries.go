package persistence

import (
	"fmt"

	"offline_twitter/scraper"
)

func (p Profile) SaveLike(l scraper.Like) error {
	_, err := p.DB.NamedExec(`
			insert into likes (sort_order, user_id, tweet_id)
			values (:sort_order, :user_id, :tweet_id)
			    on conflict do update set sort_order = max(sort_order, :sort_order)
		`,
		l,
	)
	if err != nil {
		return fmt.Errorf("Error executing SaveLike(%#v):\n  %w", l, err)
	}
	return nil
}

func (p Profile) GetLikeBySortID(id scraper.LikeSortID) (scraper.Like, error) {
	var l scraper.Like
	err := p.DB.Get(&l, `
		select sort_order, user_id, tweet_id
		  from likes
		 where sort_order = ?
	`, id)
	if err != nil {
		return l, fmt.Errorf("Error executing GetLikeBySortID(%d):\n  %w", id, err)
	}
	return l, nil
}
