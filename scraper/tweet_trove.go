package scraper

type TweetTrove struct {
	Tweets    map[TweetID]Tweet
	Users     map[UserID]User
	Retweets  map[TweetID]Retweet
}

func NewTweetTrove() TweetTrove {
	ret := TweetTrove{}
	ret.Tweets = make(map[TweetID]Tweet)
	ret.Users = make(map[UserID]User)
	ret.Retweets = make(map[TweetID]Retweet)
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
