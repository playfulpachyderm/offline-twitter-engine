package scraper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "offline_twitter/scraper"
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

	trove2 := NewTweetTrove()
	trove2.Tweets[3] = t3

	trove2.Users[1] = u1
	trove2.Users[2] = u2

	trove2.Retweets[2] = r2
	trove2.Retweets[3] = r3


	trove1.MergeWith(trove2)

	assert.Equal(3, len(trove1.Tweets))
	assert.Equal(2, len(trove1.Users))
	assert.Equal(3, len(trove1.Retweets))
}
