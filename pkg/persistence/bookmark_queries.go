package persistence

import (
	"fmt"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (p Profile) SaveBookmark(l scraper.Bookmark) error {
	_, err := p.DB.NamedExec(`
			insert into bookmarks (sort_order, user_id, tweet_id)
			values (:sort_order, :user_id, :tweet_id)
			    on conflict do update set sort_order = max(sort_order, :sort_order)
		`,
		l,
	)
	if err != nil {
		return fmt.Errorf("Error executing SaveBookmark(%#v):\n  %w", l, err)
	}
	return nil
}

func (p Profile) DeleteBookmark(l scraper.Bookmark) error {
	_, err := p.DB.NamedExec(`delete from bookmarks where user_id = :user_id and tweet_id = :tweet_id`, l)
	if err != nil {
		return fmt.Errorf("Error executing DeleteBookmark(%#v):\n  %w", l, err)
	}
	return nil
}

func (p Profile) GetBookmarkBySortID(id scraper.BookmarkSortID) (scraper.Bookmark, error) {
	var l scraper.Bookmark
	err := p.DB.Get(&l, `
		select sort_order, user_id, tweet_id
		  from bookmarks
		 where sort_order = ?
	`, id)
	if err != nil {
		return l, fmt.Errorf("Error executing GetBookmarkBySortID(%d):\n  %w", id, err)
	}
	return l, nil
}
