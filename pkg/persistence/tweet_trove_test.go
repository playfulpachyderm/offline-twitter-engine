package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestMergeTweetTroves(t *testing.T) {
	assert := assert.New(t)
	t1 := Tweet{Text: "1"}
	t2 := Tweet{Text: "2"}
	t3 := Tweet{Text: "3"}

	u1 := User{Handle: "1"}
	u2 := User{Handle: "2"}

	r1 := Retweet{TweetID: 1}
	r2 := Retweet{TweetID: 2}
	r3 := Retweet{TweetID: 3}

	trove1 := NewTweetTrove()
	trove1.Tweets[1] = t1
	trove1.Tweets[2] = t2

	trove1.Retweets[1] = r1

	trove1.TombstoneUsers = []UserHandle{"a", "b"}

	trove2 := NewTweetTrove()
	trove2.Tweets[3] = t3

	trove2.Users[1] = u1
	trove2.Users[2] = u2

	trove2.Retweets[2] = r2
	trove2.Retweets[3] = r3

	trove2.TombstoneUsers = []UserHandle{"c"}

	trove1.MergeWith(trove2)

	assert.Equal(3, len(trove1.Tweets))
	assert.Equal(2, len(trove1.Users))
	assert.Equal(3, len(trove1.Retweets))
	assert.Equal(3, len(trove1.TombstoneUsers))
}

func TestFillMissingUserIDs(t *testing.T) {
	assert := assert.New(t)
	u1 := User{ID: 1, Handle: "A"}

	t1 := Tweet{ID: 1, UserID: 1}
	t2 := Tweet{ID: 2, UserHandle: "a"}

	trove := NewTweetTrove()
	trove.Users[u1.ID] = u1
	trove.Tweets[t1.ID] = t1
	trove.Tweets[t2.ID] = t2

	assert.NotEqual(trove.Tweets[2].UserID, UserID(1))

	trove.FillMissingUserIDs()

	assert.Equal(trove.Tweets[2].UserID, UserID(1))
}

func TestFindUserByHandle(t *testing.T) {
	assert := assert.New(t)

	u1 := User{ID: 1, Handle: "1", DisplayName: "One"}
	u2 := User{ID: 2, Handle: "b", DisplayName: "Two"}

	trove := NewTweetTrove()
	trove.Users[u1.ID] = u1
	trove.Users[u2.ID] = u2

	user_2, ok := trove.FindUserByHandle("B")
	require.True(t, ok)
	assert.Equal(user_2.DisplayName, "Two")

	_, ok = trove.FindUserByHandle("3")
	assert.False(ok)
}
