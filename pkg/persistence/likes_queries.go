package persistence

import (
	"fmt"
)

func (p Profile) SaveLike(l Like) error {
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

func (p Profile) DeleteLike(l Like) error {
	_, err := p.DB.NamedExec(`delete from likes where user_id = :user_id and tweet_id = :tweet_id`, l)
	if err != nil {
		return fmt.Errorf("Error executing DeleteLike(%#v):\n  %w", l, err)
	}
	return nil
}

func (p Profile) GetLikeBySortID(id LikeSortID) (Like, error) {
	var l Like
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
