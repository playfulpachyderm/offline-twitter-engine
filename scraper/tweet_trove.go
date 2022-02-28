package scraper

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
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
 * Search for a user by handle.  Second param is whether the user was found or not.
 */
func (trove TweetTrove) FindUserByHandle(handle UserHandle) (User, bool) {
	for _, user := range trove.Users {
		if strings.EqualFold(string(user.Handle), string(handle)) {
			return user, true
		}
	}
	return User{}, false
}

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
 * Tries to fetch every User that's been identified in a tombstone in this trove
 */
func (trove *TweetTrove) FetchTombstoneUsers() {
	for _, handle := range trove.TombstoneUsers {
		// Skip fetching if this user is already in the trove
		_, already_fetched := trove.FindUserByHandle(handle)
		if already_fetched {
			continue
		}

		log.Debug("Getting tombstone user: " + handle)
		user, err := GetUser(handle)
		if err != nil {
			panic(fmt.Sprintf("Error getting tombstoned user: %s\n  %s", handle, err.Error()))
		}

		if user.ID == 0 {
			// Find some random ID to fit it into the trove
			for i := 1; ; i++ {
				_, ok := trove.Users[UserID(i)]
				if !ok {
					user.ID = UserID(i)
					break
				}
			}
		}

		trove.Users[user.ID] = user
	}
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
		if tweet.UserHandle == "" {
			// No need to fill this tweet's user_id, it's already filled
			continue
		}

		user, is_found := trove.FindUserByHandle(tweet.UserHandle)
		if !is_found {
			// The user probably deleted deleted their account, and thus `scraper.GetUser` failed.  So
			// they're not in this trove's Users.
			panic(fmt.Sprintf("Couldn't fill out this Tweet's UserID: %d, %s", tweet.ID, tweet.UserHandle))
		}
		tweet.UserID = user.ID
		trove.Tweets[i] = tweet
	}
}
