package scraper

import (
	"time"
	"fmt"
)

const DEFAULT_MAX_REPLIES_EAGER_LOAD = 50

type TweetID string

type Tweet struct {
	ID             TweetID
	UserID         UserID
	User           *User
	Text           string
	PostedAt       time.Time
	NumLikes       int
	NumRetweets    int
	NumReplies     int
	NumQuoteTweets int
	HasVideo       bool
	InReplyTo      TweetID

	Urls        []string
	Images      []string
	Mentions    []UserHandle
	Hashtags    []string
	QuotedTweet TweetID
}

func (t Tweet) String() string {
	return fmt.Sprintf(
`ID %s, User %s: %q (%s). Likes: %d, Retweets: %d, QTs: %d, Replies: %d.
Urls: %v   Images: %v   Mentions: %v   Hashtags: %v`,
	t.ID, t.UserID, t.Text, t.PostedAt, t.NumLikes, t.NumRetweets, t.NumQuoteTweets, t.NumReplies, t.Urls, t.Images, t.Mentions, t.Hashtags)
}

// Turn an APITweet, as returned from the scraper, into a properly structured Tweet object
func ParseSingleTweet(apiTweet APITweet) (ret Tweet, err error) {
	apiTweet.NormalizeContent()

	ret.ID = TweetID(apiTweet.ID)
	ret.UserID = UserID(apiTweet.UserIDStr)
	ret.Text = apiTweet.FullText

	ret.PostedAt, err = time.Parse(time.RubyDate, apiTweet.CreatedAt)
	if err != nil {
		return
	}
	ret.NumLikes = apiTweet.FavoriteCount
	ret.NumRetweets = apiTweet.RetweetCount
	ret.NumReplies = apiTweet.ReplyCount
	ret.NumQuoteTweets = apiTweet.QuoteCount
	ret.InReplyTo = TweetID(apiTweet.InReplyToStatusIDStr)

	for _, url := range apiTweet.Entities.URLs {
		ret.Urls = append(ret.Urls, url.ExpandedURL)
	}
	for _, media := range apiTweet.Entities.Media {
		if media.Type != "photo" {
			panic_str := fmt.Sprintf("Unknown media type: %q", media.Type)
			panic(panic_str)
		}
		ret.Images = append(ret.Images, media.MediaURLHttps)
	}
	for _, hashtag := range apiTweet.Entities.Hashtags {
		ret.Hashtags = append(ret.Hashtags, hashtag.Text)
	}
	for _, mention := range apiTweet.Entities.Mentions {
		ret.Mentions = append(ret.Mentions, UserHandle(mention.UserName))
	}

	ret.QuotedTweet = TweetID(apiTweet.QuotedStatusIDStr)
	ret.HasVideo = false  // TODO
	return
}


// Return a single tweet, nothing else
func GetTweet(id TweetID) (Tweet, error) {
	api := API{}
	tweet_response, err := api.GetTweet(id, "")
	if err != nil {
		return Tweet{}, fmt.Errorf("Error in API call: %s", err)
	}

	single_tweet, ok := tweet_response.GlobalObjects.Tweets[string(id)]

	if !ok {
		return Tweet{}, fmt.Errorf("Didn't get the tweet!\n%v", tweet_response)
	}

	return ParseSingleTweet(single_tweet)
}


// Return a list of tweets, including the original and the rest of its thread,
// along with a list of associated users
func GetTweetFull(id TweetID) (tweets []Tweet, retweets []Retweet, users []User, err error) {
	api := API{}
	tweet_response, err := api.GetTweet(id, "")
	if err != nil {
		return
	}
	if len(tweet_response.GlobalObjects.Tweets) < DEFAULT_MAX_REPLIES_EAGER_LOAD &&
			tweet_response.GetCursor() != "" {
		err = api.GetMoreReplies(id, &tweet_response, DEFAULT_MAX_REPLIES_EAGER_LOAD)
		if err != nil {
			return
		}
	}

	return ParseTweetResponse(tweet_response)
}

func ParseTweetResponse(resp TweetResponse) (tweets []Tweet, retweets []Retweet, users []User, err error) {
	var new_tweet Tweet
	var new_retweet Retweet
	for _, single_tweet := range resp.GlobalObjects.Tweets {
		if single_tweet.RetweetedStatusIDStr == "" {
			new_tweet, err = ParseSingleTweet(single_tweet)
			if err != nil {
				return
			}
			tweets = append(tweets, new_tweet)
		} else {
			new_retweet, err = ParseSingleRetweet(single_tweet)
			if err != nil {
				return
			}
			retweets = append(retweets, new_retweet)
		}
	}
	var new_user User
	for _, user := range resp.GlobalObjects.Users {
		new_user, err = ParseSingleUser(user)
		if err != nil {
			return
		}
		users = append(users, new_user)
	}
	return
}
