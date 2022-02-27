package persistence

import (
	"fmt"

	. "offline_twitter/scraper"
)

/**
 * Convenience function that saves all the objects in a TweetTrove.
 * Panics if anything goes wrong.
 */
func (p Profile) SaveTweetTrove(trove TweetTrove) {
	for _, u := range trove.Users {
		// Download download their tiny profile image
		err := p.DownloadUserProfileImageTiny(&u)
		if err != nil {
			panic(fmt.Sprintf("Error downloading user content for user with ID %d and handle %s: %s", u.ID, u.Handle, err.Error()))
		}

		err = p.SaveUser(&u)
		if err != nil {
			panic(fmt.Sprintf("Error saving user with ID %d and handle %s: %s", u.ID, u.Handle, err.Error()))
		}
	}

	for _, t := range trove.Tweets {
		err := p.SaveTweet(t)
		if err != nil {
			panic(fmt.Sprintf("Error saving tweet ID %d: %s", t.ID, err.Error()))
		}

		err = p.DownloadTweetContentFor(&t)
		if err != nil {
			panic(fmt.Sprintf("Error downloading tweet content for tweet ID %d: %s", t.ID, err.Error()))
		}
	}

	for _, r := range trove.Retweets {
		err := p.SaveRetweet(r)
		if err != nil {
			panic(fmt.Sprintf("Error saving retweet with ID %d from user ID %d: %s", r.RetweetID, r.RetweetedByID, err.Error()))
		}
	}
}
