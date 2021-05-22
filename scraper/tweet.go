package scraper

import (
	"time"
	"fmt"
	"strings"
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
	Mentions    []UserID
	Hashtags    []string
	QuotedTweet TweetID
}

func (t Tweet) String() string {
	return fmt.Sprintf(
`ID %s, User %s: %q (%s). Likes: %d, Retweets: %d, QTs: %d, Replies: %d.
Urls: %v   Images: %v   Mentions: %v   Hashtags: %v`,
	t.ID, t.User, t.Text, t.PostedAt, t.NumLikes, t.NumRetweets, t.NumQuoteTweets, t.NumReplies, t.Urls, t.Images, t.Mentions, t.Hashtags)
}

func ParseSingleTweet(apiTweet APITweet) (ret Tweet, err error) {
	ret.ID = TweetID(apiTweet.ID)
	ret.User = UserID(apiTweet.UserIDStr)
	ret.Text = apiTweet.FullText

	// Remove embedded links at the end of the text
	if len(apiTweet.Entities.URLs) == 1 {
		url := apiTweet.Entities.URLs[0].URL
		if strings.Index(ret.Text, url) == len(ret.Text) - len(url) {
			ret.Text = ret.Text[0:len(ret.Text) - len(url) - 1]  // Also strip the newline
		}
	}
	if len(apiTweet.Entities.Media) == 1 {
		url := apiTweet.Entities.Media[0].URL
		if strings.Index(ret.Text, url) == len(ret.Text) - len(url) {
			ret.Text = ret.Text[0:len(ret.Text) - len(url) - 1]  // Also strip the trailing space
		}
	}

	// Remove leading `@username` for replies
	if apiTweet.InReplyToScreenName != "" {
		if strings.Index(ret.Text, "@" + apiTweet.InReplyToScreenName) == 0 {
			ret.Text = ret.Text[len(apiTweet.InReplyToScreenName) + 2:]  // `@`, username, space
		}
	}

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
		ret.Mentions = append(ret.Mentions, UserID(mention.UserID))
	}

	ret.QuotedTweet = TweetID(apiTweet.QuotedStatusIDStr)
	ret.HasVideo = false  // TODO
	return
}
