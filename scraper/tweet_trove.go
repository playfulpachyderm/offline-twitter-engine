package scraper

import (
	"fmt"
)

type TweetTrove struct {
	Tweets    map[TweetID]Tweet
	Users     map[UserID]User
	Retweets  map[TweetID]Retweet

	TombstoneUsers []UserHandle
}

func NewTweetTrove() TweetTrove {
	ret := TweetTrove{}
	ret.Tweets = make(map[TweetID]Tweet)
	ret.Users = make(map[UserID]User)
	ret.Retweets = make(map[TweetID]Retweet)
	ret.TombstoneUsers = []UserHandle{}
	return ret
}

/**
 * Make it compatible with previous silly interface if needed
 */
func (trove TweetTrove) Transform() (tweets []Tweet, retweets []Retweet, users []User) {
	for _, val := range trove.Tweets {
		tweets = append(tweets, val)
	}
	for _, val := range trove.Users {
		users = append(users, val)
	}
	for _, val := range trove.Retweets {
		retweets = append(retweets, val)
	}
	return
}  // TODO: refactor until this function isn't needed anymore


/**
 * Combine two troves into one
 */
func (t1 *TweetTrove) MergeWith(t2 TweetTrove) {
	for id, val := range t2.Tweets {
		t1.Tweets[id] = val
	}
	for id, val := range t2.Users {
		t1.Users[id] = val
	}
	for id, val := range t2.Retweets {
		t1.Retweets[id] = val
	}

	t1.TombstoneUsers = append(t1.TombstoneUsers, t2.TombstoneUsers...)
}

/**
 * Checks for tombstoned tweets and fills in their UserIDs based on the collected tombstoned users.

 * To be called after calling "scraper.GetUser" on all the tombstoned users.
 *
 * At this point, those users should have been added to this trove's Users collection, and the
 * Tweets have a field `UserHandle` which can be used to pair them with newly fetched Users.
 *
 * This will still fail if the user deleted their account (instead of getting banned, blocking the
 * quote-tweeter, etc), because then that user won't show up .
 */
func (trove *TweetTrove) FillMissingUserIDs() {
	for i := range trove.Tweets {
		tweet := trove.Tweets[i]
		if tweet.UserID != 0 {
			// No need to fill this tweet's user_id, it's already filled
			continue
		}

		handle := tweet.UserHandle
		is_user_found := false
		for _, u := range trove.Users {
			if u.Handle == handle {
				tweet.UserID = u.ID
				is_user_found = true
				break
			}
		}
		if !is_user_found {
			// The user probably deleted deleted their account, and thus `scraper.GetUser` failed.  So
			// they're not in this trove's Users.
			panic(fmt.Sprintf("Couldn't fill out this Tweet's UserID: %d, %s", tweet.ID, tweet.UserHandle))
		}
		trove.Tweets[i] = tweet
	}
}
