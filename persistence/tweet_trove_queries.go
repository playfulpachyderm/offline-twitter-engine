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
	for i, u := range trove.Users {
		err := p.SaveUser(&u)
		if err != nil {
			panic(fmt.Errorf("Error saving user with ID %d and handle %s:\n  %w", u.ID, u.Handle, err))
		}
		fmt.Println(u.Handle, u.ID)
		// If the User's ID was updated in saving (i.e., Unknown User), update it in the Trove too
		// Also update tweets, retweets and spaces that reference this UserID
		for j, tweet := range trove.Tweets {
			if tweet.UserID == trove.Users[i].ID {
				tweet.UserID = u.ID
				trove.Tweets[j] = tweet
			}
		}
		for j, retweet := range trove.Retweets {
			if retweet.RetweetedByID == trove.Users[i].ID {
				retweet.RetweetedByID = u.ID
				trove.Retweets[j] = retweet
			}
		}
		trove.Users[i] = u

		// Download their tiny profile image
		err = p.DownloadUserProfileImageTiny(&u)
		if err != nil {
			panic(fmt.Errorf("Error downloading user content for user with ID %d and handle %s:\n  %w", u.ID, u.Handle, err))
		}
	}

	// TODO: this is called earlier in the process as well, before parsing.  Is that call redundant?  Too tired to figure out right now
	// Update: Yes it's redundant.  Places that return tweet troves should call `PostProcess`
	// before returning, which includes `FillMissingUserIDs`.
	// trove.FillMissingUserIDs()

	for _, t := range trove.Tweets {
		err := p.SaveTweet(t)
		if err != nil {
			panic(fmt.Errorf("Error saving tweet ID %d:\n  %w", t.ID, err))
		}

		err = p.DownloadTweetContentFor(&t)
		if err != nil {
			panic(fmt.Errorf("Error downloading tweet content for tweet ID %d:\n  %w", t.ID, err))
		}
	}

	for _, r := range trove.Retweets {
		err := p.SaveRetweet(r)
		if err != nil {
			panic(fmt.Errorf("Error saving retweet with ID %d from user ID %d:\n  %w", r.RetweetID, r.RetweetedByID, err))
		}
	}
}
