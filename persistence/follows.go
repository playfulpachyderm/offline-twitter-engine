package persistence

import (
	"offline_twitter/scraper"
)


/**
 * Create a Type for this to make it easier to expand later
 */
type Follow struct {
	Handle scraper.UserHandle  `yaml:"user"`
	AutoFetch bool             `yaml:"auto_fetch_tweets"`
}

func (p Profile) IsFollowing(handle scraper.UserHandle) bool {
	for _, follow := range p.UsersList {
		if follow.Handle == handle {
			return true;
		}
	}
	return false;
}
