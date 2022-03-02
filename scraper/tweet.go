package scraper

import (
	"time"
	"fmt"
	"strings"

	"offline_twitter/terminal_utils"
)

const DEFAULT_MAX_REPLIES_EAGER_LOAD = 50

type TweetID int64

type Tweet struct {
	ID               TweetID
	UserID           UserID
	UserHandle       UserHandle  // For processing tombstones
	User             *User
	Text             string
	PostedAt         time.Time
	NumLikes         int
	NumRetweets      int
	NumReplies       int
	NumQuoteTweets   int
	InReplyToID      TweetID
	QuotedTweetID    TweetID

	Images        []Image
	Videos        []Video
	Mentions      []UserHandle
	ReplyMentions []UserHandle
	Hashtags      []string
	Urls          []Url
	Polls         []Poll

	TombstoneType string
	IsStub bool

	IsContentDownloaded bool
	IsConversationScraped bool
	LastScrapedAt time.Time
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
		terminal_utils.FormatDate(t.PostedAt),
		terminal_utils.WrapText(t.Text, 60),
		t.NumReplies,
		t.NumRetweets,
		t.NumQuoteTweets,
		t.NumLikes,
	)

	if len(t.Images) > 0 {
		ret += fmt.Sprintf(terminal_utils.COLOR_GREEN + "images: %d\n" + terminal_utils.COLOR_RESET, len(t.Images))
	}
	if len(t.Urls) > 0 {
		ret += "urls: [\n"
		for _, url := range(t.Urls) {
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

	// Process "posted-at" date and time
	if apiTweet.TombstoneText == "" {  // Skip time parsing for tombstones
		ret.PostedAt, err = time.Parse(time.RubyDate, apiTweet.CreatedAt)
		if err != nil {
			return
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
		if media.Type != "photo" {  // TODO: remove this eventually
			panic(fmt.Sprintf("Unknown media type: %q", media.Type))
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
		ret.Mentions = append(ret.Mentions, UserHandle(mention.UserName))
	}
	for _, mention := range strings.Split(apiTweet.Entities.ReplyMentions, " ") {
		if mention != "" {
			if mention[0] != '@' {
				panic(fmt.Sprintf("Unknown ReplyMention value: %s", apiTweet.Entities.ReplyMentions))
			}
			ret.ReplyMentions = append(ret.ReplyMentions, UserHandle(mention[1:]))
		}
	}


	// Process videos
	for _, entity := range apiTweet.ExtendedEntities.Media {
		if entity.Type != "video" && entity.Type != "animated_gif" {
			continue
		}
		if len(apiTweet.ExtendedEntities.Media) != 1 {
			panic(fmt.Sprintf("Surprising ExtendedEntities: %v", apiTweet.ExtendedEntities.Media))
		}
		new_video := ParseAPIVideo(apiTweet.ExtendedEntities.Media[0], ret.ID)
		ret.Videos = []Video{new_video}
		ret.Images = []Image{}
	}

	// Process polls
	if strings.Index(apiTweet.Card.Name, "poll") == 0 {
		poll := ParseAPIPoll(apiTweet.Card)
		poll.TweetID = ret.ID
		ret.Polls = []Poll{poll}
	}

	// Process tombstones and other metadata
	ret.TombstoneType = apiTweet.TombstoneText
	ret.IsStub = !(ret.TombstoneType == "")
	ret.LastScrapedAt = time.Unix(0, 0)  // Caller will change this for the tweet that was actually scraped
	ret.IsConversationScraped = false  // Safe due to the "No Worsening" principle

	return
}


/**
 * Get a single tweet with no replies from the API.
 *
 * args:
 * - id: the ID of the tweet to get
 *
 * returns: the single Tweet
 */
func GetTweet(id TweetID) (Tweet, error) {
	api := API{}
	tweet_response, err := api.GetTweet(id, "")
	if err != nil {
		return Tweet{}, fmt.Errorf("Error in API call: %s", err)
	}

	single_tweet, ok := tweet_response.GlobalObjects.Tweets[fmt.Sprint(id)]

	if !ok {
		return Tweet{}, fmt.Errorf("Didn't get the tweet!\n%v", tweet_response)
	}

	return ParseSingleTweet(single_tweet)
}


/**
 * Return a list of tweets, including the original and the rest of its thread,
 * along with a list of associated users.
 *
 * Mark the main tweet as "is_conversation_downloaded = true", and update its "last_scraped_at"
 * value.
 *
 * args:
 * - id: the ID of the tweet to get
 *
 * returns: the tweet, list of its replies and context, and users associated with those replies
 */
func GetTweetFull(id TweetID) (trove TweetTrove, err error) {
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

	// This has to be called BEFORE ParseTweetResponse, because it modifies the TweetResponse (adds tombstone tweets to its tweets list)
	tombstoned_users := tweet_response.HandleTombstones()

	trove, err = ParseTweetResponse(tweet_response)
	if err != nil {
		panic(err)
	}
	trove.TombstoneUsers = tombstoned_users
	trove.FetchTombstoneUsers()

	// Quoted tombstones need their user_id filled out from the tombstoned_users list
	trove.FillMissingUserIDs()

	// Find the main tweet and update its "is_conversation_downloaded" and "last_scraped_at"
	tweet, ok := trove.Tweets[id]
	if !ok {
		panic("Trove didn't contain its own tweet!")
	}
	tweet.LastScrapedAt = time.Now()
	tweet.IsConversationScraped = true
	trove.Tweets[id] = tweet

	// tweets, retweets, users = trove.Transform()

	return
}

/**
 * Parse an API response object into a list of tweets, retweets and users
 *
 * args:
 * - resp: the response from the API
 *
 * returns: a list of tweets, retweets and users in that response object
 */
func ParseTweetResponse(resp TweetResponse) (TweetTrove, error) {
	trove := NewTweetTrove()

	for _, single_tweet := range resp.GlobalObjects.Tweets {
		if single_tweet.RetweetedStatusIDStr == "" {
			new_tweet, err := ParseSingleTweet(single_tweet)
			if err != nil {
				return trove, err
			}
			trove.Tweets[new_tweet.ID] = new_tweet
		} else {
			new_retweet, err := ParseSingleRetweet(single_tweet)
			if err != nil {
				return trove, err
			}
			trove.Retweets[new_retweet.RetweetID] = new_retweet
		}
	}

	for _, user := range resp.GlobalObjects.Users {
		new_user, err := ParseSingleUser(user)
		if err != nil {
			return trove, err
		}
		trove.Users[new_user.ID] = new_user
	}
	return trove, nil
}
