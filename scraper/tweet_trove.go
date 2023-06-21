package scraper

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type TweetTrove struct {
	Tweets   map[TweetID]Tweet
	Users    map[UserID]User
	Retweets map[TweetID]Retweet
	Spaces   map[SpaceID]Space

	TombstoneUsers []UserHandle
}

func NewTweetTrove() TweetTrove {
	ret := TweetTrove{}
	ret.Tweets = make(map[TweetID]Tweet)
	ret.Users = make(map[UserID]User)
	ret.Retweets = make(map[TweetID]Retweet)
	ret.Spaces = make(map[SpaceID]Space)
	ret.TombstoneUsers = []UserHandle{}
	return ret
}

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
	for id, val := range t2.Spaces {
		t1.Spaces[id] = val
	}

	t1.TombstoneUsers = append(t1.TombstoneUsers, t2.TombstoneUsers...)
}

/**
 * Tries to fetch every User that's been identified in a tombstone in this trove
 */
func (trove *TweetTrove) FetchTombstoneUsers() {
	for _, handle := range trove.TombstoneUsers {
		// Skip fetching if this user is already in the trove
		user, already_fetched := trove.FindUserByHandle(handle)

		if already_fetched {
			// If the user is already fetched and it's an intact user, don't fetch it again
			if user.JoinDate.Unix() != (Timestamp{}).Unix() {
				log.Debugf("Skipping %q due to intact user", handle)
				continue
			}

			// A user needs a valid handle or ID to fetch it by
			if user.IsIdFake && user.Handle == "<UNKNOWN USER>" {
				log.Debugf("Skipping %q due to completely unknown user (not fetchable)", handle)
				continue
			}
		}

		log.Debug("Getting tombstone user: " + handle)
		user, err := GetUser(handle)
		if err != nil {
			panic(fmt.Errorf("Error getting tombstoned user with handle %q: \n  %w", handle, err))
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
			panic(fmt.Errorf(
				"Couldn't find user ID for user %q, while filling missing UserID in tweet with ID %d",
				tweet.UserHandle,
				tweet.ID,
			))
		}
		tweet.UserID = user.ID
		trove.Tweets[i] = tweet
	}
}

func (trove *TweetTrove) FillSpaceDetails() error {
	fmt.Println("Filling space details")
	for i := range trove.Spaces {
		fmt.Printf("Getting space: %q\n", trove.Spaces[i].ID)
		new_trove, err := FetchSpaceDetail(trove.Spaces[i].ID)
		if err != nil {
			return err
		}
		// Replace the old space in the trove with the new, updated one
		new_space, is_ok := new_trove.Spaces[i]
		if new_space.ShortUrl == "" {
			// Copy over the short-url, which doesn't seem to exist on a full Space response
			new_space.ShortUrl = trove.Spaces[i].ShortUrl
		}
		if is_ok {
			// Necessary to check is_ok because the space response could be empty, in which case
			// we don't want to overwrite it
			trove.Spaces[i] = new_space
		}
	}
	return nil
}

func (trove *TweetTrove) PostProcess() error {
	trove.FetchTombstoneUsers()
	trove.FillMissingUserIDs()
	err := trove.FillSpaceDetails()
	if err != nil {
		return err
	}
	return nil
}
