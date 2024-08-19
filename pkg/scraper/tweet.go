package scraper

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/terminal_utils"
)

var ERR_NO_TWEET = errors.New("Empty tweet")

type TweetID int64

type CommaSeparatedList []string

func (l *CommaSeparatedList) Scan(src interface{}) error {
	*l = CommaSeparatedList{}
	switch src := src.(type) {
	case string:
		for _, v := range strings.Split(src, ",") {
			if v != "" {
				*l = append(*l, v)
			}
		}
	default:
		panic("Should be a string")
	}
	return nil
}
func (l CommaSeparatedList) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}

type Tweet struct {
	ID             TweetID   `db:"id"`
	Text           string    `db:"text"`
	IsExpandable   bool      `db:"is_expandable"`
	PostedAt       Timestamp `db:"posted_at"`
	NumLikes       int       `db:"num_likes"`
	NumRetweets    int       `db:"num_retweets"`
	NumReplies     int       `db:"num_replies"`
	NumQuoteTweets int       `db:"num_quote_tweets"`
	InReplyToID    TweetID   `db:"in_reply_to_id"`
	QuotedTweetID  TweetID   `db:"quoted_tweet_id"`

	UserID UserID `db:"user_id"`
	User   *User  `db:"user"`

	// For processing tombstones
	UserHandle              UserHandle
	in_reply_to_user_handle UserHandle
	in_reply_to_user_id     UserID

	Images        []Image
	Videos        []Video
	Urls          []Url
	Polls         []Poll
	Mentions      CommaSeparatedList `db:"mentions"`
	ReplyMentions CommaSeparatedList `db:"reply_mentions"`
	Hashtags      CommaSeparatedList `db:"hashtags"`

	// TODO get-rid-of-redundant-spaces: Might be good to get rid of `Spaces`.  Only used in APIv1 I think.
	// A first-step would be to delete the Spaces after pulling them out of a Tweet into the Trove
	// in ToTweetTrove.  Then they will only be getting saved once rather than twice.
	Spaces  []Space
	SpaceID SpaceID `db:"space_id"`

	TombstoneType string `db:"tombstone_type"`
	TombstoneText string `db:"tombstone_text"`
	IsStub        bool   `db:"is_stub"`

	IsLikedByCurrentUser  bool      `db:"is_liked_by_current_user"`
	IsContentDownloaded   bool      `db:"is_content_downloaded"`
	IsConversationScraped bool      `db:"is_conversation_scraped"`
	LastScrapedAt         Timestamp `db:"last_scraped_at"`
}

func (t Tweet) String() string {
	var author string
	if t.User != nil {
		author = fmt.Sprintf("%s\n@%s", t.User.DisplayName, t.User.Handle)
	} else {
		author = "@???"
	}

	ret := fmt.Sprintf(
		`%s
%s
%s
Replies: %d      RT: %d      QT: %d      Likes: %d
`,
		author,
		terminal_utils.FormatDate(t.PostedAt.Time),
		terminal_utils.WrapText(t.Text, 60),
		t.NumReplies,
		t.NumRetweets,
		t.NumQuoteTweets,
		t.NumLikes,
	)

	if len(t.Images) > 0 {
		ret += fmt.Sprintf(terminal_utils.COLOR_GREEN+"images: %d\n"+terminal_utils.COLOR_RESET, len(t.Images))
	}
	if len(t.Urls) > 0 {
		ret += "urls: [\n"
		for _, url := range t.Urls {
			ret += "  " + url.Text + "\n"
		}
		ret += "]"
	}

	return ret
}

// Turn an APITweet, as returned from the scraper, into a properly structured Tweet object
func ParseSingleTweet(apiTweet APITweet) (ret Tweet, err error) {
	apiTweet.NormalizeContent()

	ret.ID = TweetID(apiTweet.ID)
	ret.UserID = UserID(apiTweet.UserID)
	ret.UserHandle = UserHandle(apiTweet.UserHandle)
	ret.Text = apiTweet.FullText
	ret.IsExpandable = apiTweet.IsExpandable

	// Process "posted-at" date and time
	if apiTweet.TombstoneText == "" { // Skip time parsing for tombstones
		ret.PostedAt, err = TimestampFromString(apiTweet.CreatedAt)
		if err != nil {
			if ret.ID == 0 {
				return Tweet{}, fmt.Errorf("unable to parse tweet: %w", ERR_NO_TWEET)
			}
			return Tweet{}, fmt.Errorf("Error parsing time on tweet ID %d:\n  %w", ret.ID, err)
		}
	}

	ret.NumLikes = apiTweet.FavoriteCount
	ret.NumRetweets = apiTweet.RetweetCount
	ret.NumReplies = apiTweet.ReplyCount
	ret.NumQuoteTweets = apiTweet.QuoteCount
	ret.InReplyToID = TweetID(apiTweet.InReplyToStatusID)
	ret.QuotedTweetID = TweetID(apiTweet.QuotedStatusID)

	// Process URLs and link previews
	for _, url := range apiTweet.Entities.URLs {
		var url_object Url
		if apiTweet.Card.ShortenedUrl == url.ShortenedUrl {
			if apiTweet.Card.Name == "3691233323:audiospace" {
				// This "url" is just a link to a Space.  Don't process it as a Url
				continue
			}
			url_object = ParseAPIUrlCard(apiTweet.Card)
		}
		url_object.Text = url.ExpandedURL
		url_object.ShortText = url.ShortenedUrl
		url_object.TweetID = ret.ID

		// Skip it if it's just the quoted tweet
		_, id, is_ok := TryParseTweetUrl(url.ExpandedURL)
		if is_ok && id == ret.QuotedTweetID {
			continue
		}

		ret.Urls = append(ret.Urls, url_object)
	}

	// Process images
	for _, media := range apiTweet.Entities.Media {
		if media.Type != "photo" {
			// Videos now have an entry in "Entities.Media" but they can be ignored; the useful bit is in ExtendedEntities
			// So skip ones that aren't "photo"
			continue
		}
		new_image := ParseAPIMedia(media)
		new_image.TweetID = ret.ID
		ret.Images = append(ret.Images, new_image)
	}

	// Process hashtags
	for _, hashtag := range apiTweet.Entities.Hashtags {
		ret.Hashtags = append(ret.Hashtags, hashtag.Text)
	}

	// Process `@` mentions and reply-mentions
	for _, mention := range apiTweet.Entities.Mentions {
		ret.Mentions = append(ret.Mentions, mention.UserName)
	}
	for _, mention := range strings.Split(apiTweet.Entities.ReplyMentions, " ") {
		if mention != "" {
			if mention[0] != '@' {
				panic(fmt.Errorf("Unknown ReplyMention value %q:\n  %w", apiTweet.Entities.ReplyMentions, EXTERNAL_API_ERROR))
			}
			ret.ReplyMentions = append(ret.ReplyMentions, mention[1:])
		}
	}

	// Process videos
	for _, entity := range apiTweet.ExtendedEntities.Media {
		if entity.Type != "video" && entity.Type != "animated_gif" {
			continue
		}

		new_video := ParseAPIVideo(entity)
		new_video.TweetID = ret.ID
		ret.Videos = append(ret.Videos, new_video)

		// Remove the thumbnail from the Images list
		updated_imgs := []Image{}
		for _, img := range ret.Images {
			if VideoID(img.ID) != new_video.ID {
				updated_imgs = append(updated_imgs, img)
			}
		}
		ret.Images = updated_imgs
	}

	// Process polls
	if strings.Index(apiTweet.Card.Name, "poll") == 0 {
		poll := ParseAPIPoll(apiTweet.Card)
		poll.TweetID = ret.ID
		ret.Polls = []Poll{poll}
	}

	// Process spaces
	if apiTweet.Card.Name == "3691233323:audiospace" {
		space := ParseAPISpace(apiTweet.Card)
		ret.Spaces = []Space{space}
		ret.SpaceID = space.ID
	}

	// Process tombstones and other metadata
	ret.TombstoneType = apiTweet.TombstoneText
	ret.IsStub = !(ret.TombstoneType == "")
	ret.LastScrapedAt = TimestampFromUnix(0) // Caller will change this for the tweet that was actually scraped
	ret.IsConversationScraped = false        // Safe due to the "No Worsening" principle

	// Extra data that can help piece together tombstoned tweet info
	ret.in_reply_to_user_id = UserID(apiTweet.InReplyToUserID)
	ret.in_reply_to_user_handle = UserHandle(apiTweet.InReplyToScreenName)

	return
}

// Get a single tweet with no replies from the API.
//
// args:
// - id: the ID of the tweet to get
//
// returns: the single Tweet
func (api *API) GetTweet(id TweetID) (Tweet, error) {
	resp, err := api.GetTweetDetail(id, "")
	if err != nil {
		return Tweet{}, fmt.Errorf("Error getting tweet detail: %d\n  %w", id, err)
	}
	trove, err := resp.ToTweetTrove()
	if err != nil {
		return Tweet{}, err
	}

	// Find the main tweet and update its "is_conversation_downloaded" and "last_scraped_at"
	tweet, ok := trove.Tweets[id]
	if !ok {
		panic("Trove didn't contain its own tweet!")
	}
	tweet.LastScrapedAt = Timestamp{time.Now()}
	tweet.IsConversationScraped = true
	return tweet, nil
}
