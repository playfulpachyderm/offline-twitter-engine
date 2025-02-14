package persistence

import (
	"fmt"
	"strings"
)

type TweetTrove struct {
	Tweets        map[TweetID]Tweet
	Users         map[UserID]User
	Retweets      map[TweetID]Retweet
	Spaces        map[SpaceID]Space
	Likes         map[LikeSortID]Like
	Bookmarks     map[BookmarkSortID]Bookmark
	Notifications map[NotificationID]Notification

	TombstoneUsers []UserHandle

	// For DMs
	Rooms    map[DMChatRoomID]DMChatRoom
	Messages map[DMMessageID]DMMessage
}

func NewTweetTrove() TweetTrove {
	ret := TweetTrove{}
	ret.Tweets = make(map[TweetID]Tweet)
	ret.Users = make(map[UserID]User)
	ret.Retweets = make(map[TweetID]Retweet)
	ret.Spaces = make(map[SpaceID]Space)
	ret.Likes = make(map[LikeSortID]Like)
	ret.Bookmarks = make(map[BookmarkSortID]Bookmark)
	ret.Notifications = make(map[NotificationID]Notification)
	ret.TombstoneUsers = []UserHandle{}
	ret.Rooms = make(map[DMChatRoomID]DMChatRoom)
	ret.Messages = make(map[DMMessageID]DMMessage)
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
	for id, val := range t2.Likes {
		t1.Likes[id] = val
	}
	for id, val := range t2.Bookmarks {
		t1.Bookmarks[id] = val
	}
	for id, val := range t2.Notifications {
		t1.Notifications[id] = val
	}

	t1.TombstoneUsers = append(t1.TombstoneUsers, t2.TombstoneUsers...)

	for id, val := range t2.Rooms {
		t1.Rooms[id] = val
	}
	for id, val := range t2.Messages {
		t1.Messages[id] = val
	}
}

// Checks for tombstoned tweets and fills in their UserIDs based on the collected tombstoned users.
// To be called after calling "scraper.GetUser" on all the tombstoned users.
//
// At this point, those users should have been added to this trove's Users collection, and the
// Tweets have a field `UserHandle` which can be used to pair them with newly fetched Users.
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

func (t TweetTrove) GetOldestMessage(id DMChatRoomID) DMMessageID {
	oldest := DMMessageID(^uint(0) >> 1) // Max integer
	for _, m := range t.Messages {
		if m.ID < oldest && m.DMChatRoomID == id {
			oldest = m.ID
		}
	}
	return oldest
}
