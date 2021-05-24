package scraper

import (
	"time"
	"fmt"
)


type TweetID string

type Tweet struct {
	ID             TweetID
	User           UserID
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
	Mentions    []string
	Hashtags    []string
	QuotedTweet TweetID
}

func (t Tweet) String() string {
	return fmt.Sprintf(
`ID %s, User %s: %q (%s). Likes: %d, Retweets: %d, QTs: %d, Replies: %d.
Urls: %v   Images: %v   Mentions: %v   Hashtags: %v`,
	t.ID, t.User, t.Text, t.PostedAt, t.NumLikes, t.NumRetweets, t.NumQuoteTweets, t.NumReplies, t.Urls, t.Images, t.Mentions, t.Hashtags)
}

// Turn an APITweet, as returned from the scraper, into a properly structured Tweet object
func ParseSingleTweet(apiTweet APITweet) (ret Tweet, err error) {
	apiTweet.NormalizeContent()

	ret.ID = TweetID(apiTweet.ID)
	ret.User = UserID(apiTweet.UserIDStr)
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
		ret.Mentions = append(ret.Mentions, mention.UserName)
	}

	ret.QuotedTweet = TweetID(apiTweet.QuotedStatusIDStr)
	ret.HasVideo = false  // TODO
	return
}
