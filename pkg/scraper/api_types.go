package scraper

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type APIMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	OriginalInfo  struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"original_info"`
}

func ParseAPIMedia(apiMedia APIMedia) Image {
	local_filename := get_prefixed_path(path.Base(apiMedia.MediaURLHttps))

	return Image{
		ID:            ImageID(apiMedia.ID),
		RemoteURL:     apiMedia.MediaURLHttps,
		Width:         apiMedia.OriginalInfo.Width,
		Height:        apiMedia.OriginalInfo.Height,
		LocalFilename: local_filename,
		IsDownloaded:  false,
	}
}

type SortableVariants []struct {
	Bitrate int    `json:"bitrate,omitempty"`
	URL     string `json:"url"`
}

func (v SortableVariants) Len() int           { return len(v) }
func (v SortableVariants) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v SortableVariants) Less(i, j int) bool { return v[i].Bitrate > v[j].Bitrate }

type APIExtendedMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	VideoInfo     struct {
		Variants SortableVariants `json:"variants"`
		Duration int              `json:"duration_millis"`
	} `json:"video_info"`
	ExtMediaAvailability struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"ext_media_availability"`
	OriginalInfo struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"original_info"`
	Ext struct {
		MediaStats struct {
			R interface{} `json:"r"`
		} `json:"mediaStats"`
	} `json:"ext"`
	URL string `json:"url"` // For DM videos
}

type APICard struct {
	Name          string `json:"name"`
	ShortenedUrl  string `json:"url"`
	BindingValues struct {
		Domain struct {
			Value string `json:"string_value"`
		} `json:"domain"`
		Creator struct {
			UserValue struct {
				Value int64 `json:"id_str,string"`
			} `json:"user_value"`
		} `json:"creator"`
		Site struct {
			UserValue struct {
				Value int64 `json:"id_str,string"`
			} `json:"user_value"`
		} `json:"site"`
		Title struct {
			Value string `json:"string_value"`
		} `json:"title"`
		Description struct {
			Value string `json:"string_value"`
		} `json:"description"`
		Thumbnail struct {
			ImageValue struct {
				Url    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"image_value"`
		} `json:"thumbnail_image_large"`
		PlayerImage struct {
			ImageValue struct {
				Url string `json:"url"`
			} `json:"image_value"`
		} `json:"player_image_large"`

		// For polls
		Choice1 struct {
			StringValue string `json:"string_value"`
		} `json:"choice1_label"`
		Choice2 struct {
			StringValue string `json:"string_value"`
		} `json:"choice2_label"`
		Choice3 struct {
			StringValue string `json:"string_value"`
		} `json:"choice3_label"`
		Choice4 struct {
			StringValue string `json:"string_value"`
		} `json:"choice4_label"`

		Choice1_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice1_count"`
		Choice2_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice2_count"`
		Choice3_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice3_count"`
		Choice4_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice4_count"`

		EndDatetimeUTC struct {
			StringValue string `json:"string_value"`
		} `json:"end_datetime_utc"`
		CountsAreFinal struct {
			BooleanValue bool `json:"boolean_value"`
		} `json:"counts_are_final"`
		DurationMinutes struct {
			StringValue string `json:"string_value"`
		} `json:"duration_minutes"`
		LastUpdatedAt struct {
			StringValue string `json:"string_value"`
		} `json:"last_updated_datetime_utc"`

		// For Spaces
		ID struct {
			StringValue string `json:"string_value"`
		} `json:"id"`
	} `json:"binding_values"`
}

func ParseAPIPoll(apiCard APICard) Poll {
	card_url, err := url.Parse(apiCard.ShortenedUrl)
	if err != nil {
		panic(err)
	}
	id := int_or_panic(card_url.Hostname())

	ret := Poll{}
	ret.ID = PollID(id)
	ret.NumChoices = parse_num_choices(apiCard.Name)
	ret.VotingDuration = int_or_panic(apiCard.BindingValues.DurationMinutes.StringValue) * 60
	ret.VotingEndsAt, err = TimestampFromString(apiCard.BindingValues.EndDatetimeUTC.StringValue)
	if err != nil {
		panic(err)
	}
	ret.LastUpdatedAt, err = TimestampFromString(apiCard.BindingValues.LastUpdatedAt.StringValue)
	if err != nil {
		panic(err)
	}

	ret.Choice1 = apiCard.BindingValues.Choice1.StringValue
	ret.Choice1_Votes = int_or_panic(apiCard.BindingValues.Choice1_Count.StringValue)
	ret.Choice2 = apiCard.BindingValues.Choice2.StringValue
	ret.Choice2_Votes = int_or_panic(apiCard.BindingValues.Choice2_Count.StringValue)

	if ret.NumChoices > 2 {
		ret.Choice3 = apiCard.BindingValues.Choice3.StringValue
		ret.Choice3_Votes = int_or_panic(apiCard.BindingValues.Choice3_Count.StringValue)
	}
	if ret.NumChoices > 3 {
		ret.Choice4 = apiCard.BindingValues.Choice4.StringValue
		ret.Choice4_Votes = int_or_panic(apiCard.BindingValues.Choice4_Count.StringValue)
	}

	return ret
}

func parse_num_choices(card_name string) int {
	if strings.Index(card_name, "poll") != 0 || strings.Index(card_name, "choice") != 5 {
		panic("Not valid card name: " + card_name)
	}

	return int_or_panic(card_name[4:5])
}

func ParseAPIVideo(apiVideo APIExtendedMedia) Video {
	variants := apiVideo.VideoInfo.Variants
	sort.Sort(variants)
	video_remote_url := variants[0].URL

	var view_count int

	r := apiVideo.Ext.MediaStats.R

	switch r.(type) {
	case string:
		view_count = 0
	case map[string]interface{}:
		OK_entry, ok := r.(map[string]interface{})["ok"]
		if !ok {
			panic("No 'ok' value found in the R!")
		}
		view_count_str, ok := OK_entry.(map[string]interface{})["viewCount"]
		view_count = int_or_panic(view_count_str.(string))
		if !ok {
			panic("No 'viewCount' value found in the OK!")
		}
	}

	video_parsed_url, err := url.Parse(video_remote_url)
	if err != nil {
		panic(err)
	}

	local_filename := get_prefixed_path(path.Base(video_parsed_url.Path))

	return Video{
		ID:            VideoID(apiVideo.ID),
		Width:         apiVideo.OriginalInfo.Width,
		Height:        apiVideo.OriginalInfo.Height,
		RemoteURL:     video_remote_url,
		LocalFilename: local_filename,

		ThumbnailRemoteUrl: apiVideo.MediaURLHttps,
		ThumbnailLocalPath: get_prefixed_path(path.Base(apiVideo.MediaURLHttps)),
		Duration:           apiVideo.VideoInfo.Duration,
		ViewCount:          view_count,

		IsDownloaded:    false,
		IsBlockedByDMCA: false,
		IsGeoblocked:    apiVideo.ExtMediaAvailability.Reason == "Geoblocked",
		IsGif:           apiVideo.Type == "animated_gif",
	}
}

func ParseAPIUrlCard(apiCard APICard) Url {
	values := apiCard.BindingValues
	ret := Url{}
	ret.HasCard = true

	ret.Domain = values.Domain.Value
	ret.Title = values.Title.Value
	ret.Description = values.Description.Value
	ret.IsContentDownloaded = false
	ret.CreatorID = UserID(values.Creator.UserValue.Value)
	ret.SiteID = UserID(values.Site.UserValue.Value)

	var thumbnail_url string

	if apiCard.Name == "summary_large_image" || apiCard.Name == "summary" {
		thumbnail_url = values.Thumbnail.ImageValue.Url
	} else if apiCard.Name == "player" {
		thumbnail_url = values.PlayerImage.ImageValue.Url
	} else if apiCard.Name == "unified_card" {
		// TODO: Grok chat previews
		log.Print("Grok chat card, not implemented yet-- skipping")
	} else {
		panic("Unknown card type: " + apiCard.Name)
	}

	if thumbnail_url != "" {
		ret.HasThumbnail = true
		ret.ThumbnailRemoteUrl = thumbnail_url
		ret.ThumbnailLocalPath = get_thumbnail_local_path(thumbnail_url)
		ret.ThumbnailWidth = values.Thumbnail.ImageValue.Width
		ret.ThumbnailHeight = values.Thumbnail.ImageValue.Height
	}

	return ret
}

func get_prefixed_path(p string) string {
	local_prefix_regex := regexp.MustCompile(`^[\w-]{2}`)
	local_prefix := local_prefix_regex.FindString(p)
	if len(local_prefix) != 2 {
		panic(fmt.Sprintf("Unable to extract a 2-letter prefix for filename %s", p))
	}
	return path.Join(local_prefix, p)
}

func get_thumbnail_local_path(remote_url string) string {
	u, err := url.Parse(remote_url)
	if err != nil {
		panic(err)
	}
	if u.RawQuery == "" {
		return path.Base(u.Path)
	}
	query_params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		panic(err)
	}

	return get_prefixed_path(
		fmt.Sprintf("%s_%s.%s", path.Base(u.Path), query_params["name"][0], query_params["format"][0]),
	)
}

type APITweet struct {
	ID               int64  `json:"id_str,string"`
	ConversationID   int64  `json:"conversation_id_str,string"`
	CreatedAt        string `json:"created_at"`
	FavoriteCount    int    `json:"favorite_count"`
	FullText         string `json:"full_text"`
	DisplayTextRange []int  `json:"display_text_range"`
	Entities         struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Media []APIMedia `json:"media"`
		URLs  []struct {
			ExpandedURL  string `json:"expanded_url"`
			ShortenedUrl string `json:"url"`
		} `json:"urls"`
		Mentions []struct {
			UserName string `json:"screen_name"`
			UserID   int64  `json:"id_str,string"`
		} `json:"user_mentions"`
		ReplyMentions string // The leading part of the text which is cut off by "DisplayTextRange"
	} `json:"entities"`
	ExtendedEntities struct {
		Media []APIExtendedMedia `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusID     int64  `json:"in_reply_to_status_id_str,string"`
	InReplyToUserID       int64  `json:"in_reply_to_user_id_str,string"`
	InReplyToScreenName   string `json:"in_reply_to_screen_name"`
	ReplyCount            int    `json:"reply_count"`
	RetweetCount          int    `json:"retweet_count"`
	QuoteCount            int    `json:"quote_count"`
	RetweetedStatusIDStr  string `json:"retweeted_status_id_str"` // Can be empty string
	RetweetedStatusID     int64
	QuotedStatusIDStr     string `json:"quoted_status_id_str"` // Can be empty string
	QuotedStatusID        int64
	QuotedStatusPermalink struct {
		ShortURL    string `json:"url"`
		ExpandedURL string `json:"expanded"`
	} `json:"quoted_status_permalink"`
	Time          time.Time `json:"time"`
	UserID        int64     `json:"user_id_str,string"`
	UserHandle    string
	Card          APICard `json:"card"`
	TombstoneText string
	IsExpandable  bool
}

func (t APITweet) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()
	if t.RetweetedStatusIDStr == "" {
		// Parse as a Tweet
		new_tweet, err := ParseSingleTweet(t)
		if err != nil {
			return ret, err
		}
		ret.Tweets[new_tweet.ID] = new_tweet
		for _, space := range new_tweet.Spaces {
			ret.Spaces[space.ID] = space
		}
	} else {
		// Parse as a Retweet
		new_retweet := Retweet{}
		var err error

		t.NormalizeContent()

		new_retweet.RetweetID = TweetID(t.ID)
		new_retweet.TweetID = TweetID(t.RetweetedStatusID)
		new_retweet.RetweetedByID = UserID(t.UserID)
		new_retweet.RetweetedAt, err = TimestampFromString(t.CreatedAt)
		if err != nil {
			return ret, err
		}
		ret.Retweets[new_retweet.RetweetID] = new_retweet
	}
	return ret, nil
}

// Turn an APITweet, as returned from the scraper, into a properly structured Tweet object
func ParseSingleTweet(t APITweet) (ret Tweet, err error) {
	t.NormalizeContent()

	ret.ID = TweetID(t.ID)
	ret.UserID = UserID(t.UserID)
	ret.UserHandle = UserHandle(t.UserHandle)
	ret.Text = t.FullText
	ret.IsExpandable = t.IsExpandable

	// Process "posted-at" date and time
	if t.TombstoneText == "" { // Skip time parsing for tombstones
		ret.PostedAt, err = TimestampFromString(t.CreatedAt)
		if err != nil {
			if ret.ID == 0 {
				return Tweet{}, fmt.Errorf("unable to parse tweet: %w", ERR_NO_TWEET)
			}
			return Tweet{}, fmt.Errorf("Error parsing time on tweet ID %d:\n  %w", ret.ID, err)
		}
	}

	ret.NumLikes = t.FavoriteCount
	ret.NumRetweets = t.RetweetCount
	ret.NumReplies = t.ReplyCount
	ret.NumQuoteTweets = t.QuoteCount
	ret.InReplyToID = TweetID(t.InReplyToStatusID)
	ret.QuotedTweetID = TweetID(t.QuotedStatusID)

	// Process URLs and link previews
	for _, url := range t.Entities.URLs {
		var url_object Url
		if t.Card.ShortenedUrl == url.ShortenedUrl {
			if t.Card.Name == "3691233323:audiospace" {
				// This "url" is just a link to a Space.  Don't process it as a Url
				continue
			}
			url_object = ParseAPIUrlCard(t.Card)
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
	for _, media := range t.Entities.Media {
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
	for _, hashtag := range t.Entities.Hashtags {
		ret.Hashtags = append(ret.Hashtags, hashtag.Text)
	}

	// Process `@` mentions and reply-mentions
	for _, mention := range t.Entities.Mentions {
		ret.Mentions = append(ret.Mentions, mention.UserName)
	}
	for _, mention := range strings.Split(t.Entities.ReplyMentions, " ") {
		if mention != "" {
			if mention[0] != '@' {
				panic(fmt.Errorf("Unknown ReplyMention value %q:\n  %w", t.Entities.ReplyMentions, EXTERNAL_API_ERROR))
			}
			ret.ReplyMentions = append(ret.ReplyMentions, mention[1:])
		}
	}

	// Process videos
	for _, entity := range t.ExtendedEntities.Media {
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
	if strings.Index(t.Card.Name, "poll") == 0 {
		poll := ParseAPIPoll(t.Card)
		poll.TweetID = ret.ID
		ret.Polls = []Poll{poll}
	}

	// Process spaces
	if t.Card.Name == "3691233323:audiospace" {
		space := Space{}
		space.ID = SpaceID(t.Card.BindingValues.ID.StringValue)
		space.ShortUrl = t.Card.ShortenedUrl

		// Indicate that this Space needs its details fetched still
		space.IsDetailsFetched = false

		ret.Spaces = []Space{space}
		ret.SpaceID = space.ID
	}

	// Process tombstones and other metadata
	ret.TombstoneType = t.TombstoneText
	ret.IsStub = !(ret.TombstoneType == "")
	ret.LastScrapedAt = TimestampFromUnix(0) // Caller will change this for the tweet that was actually scraped
	ret.IsConversationScraped = false        // Safe due to the "No Worsening" principle

	// Extra data that can help piece together tombstoned tweet info
	ret.in_reply_to_user_id = UserID(t.InReplyToUserID)
	ret.in_reply_to_user_handle = UserHandle(t.InReplyToScreenName)

	return
}

func (t *APITweet) NormalizeContent() {
	id, err := strconv.Atoi(t.QuotedStatusIDStr)
	if err == nil {
		t.QuotedStatusID = int64(id)
	}
	id, err = strconv.Atoi(t.RetweetedStatusIDStr)
	if err == nil {
		t.RetweetedStatusID = int64(id)
	}

	if len(t.DisplayTextRange) == 2 {
		t.Entities.ReplyMentions = strings.TrimSpace(string([]rune(t.FullText)[0:t.DisplayTextRange[0]]))
		t.FullText = string([]rune(t.FullText)[t.DisplayTextRange[0]:t.DisplayTextRange[1]])
	}

	// Handle short links showing up at ends of tweets
	for _, url := range t.Entities.URLs {
		index := strings.Index(t.FullText, url.ShortenedUrl)
		if index < 0 {
			// It's not in the text
			continue
		}
		if index == (len(t.FullText) - len(url.ShortenedUrl)) {
			t.FullText = strings.TrimSpace(t.FullText[0:index])
		}
	}

	// Handle pasted tweet links that turn into quote tweets but still have a link in them
	// This is a separate case from above because we want it gone even if it's in the middle of the tweet
	if t.QuotedStatusID != 0 {
		for _, url := range t.Entities.URLs {
			if url.ShortenedUrl == t.QuotedStatusPermalink.ShortURL {
				t.FullText = strings.ReplaceAll(t.FullText, url.ShortenedUrl, "")
			}
		}
	}
	t.FullText = html.UnescapeString(t.FullText)
	t.FullText = strings.TrimSpace(t.FullText)
}

func (t APITweet) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(data)
}

type APIUser struct {
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Entities    struct {
		URL struct {
			Urls []struct {
				ExpandedURL string `json:"expanded_url"`
			} `json:"urls"`
		} `json:"url"`
	} `json:"entities"`
	FavouritesCount      int      `json:"favourites_count"`
	FollowersCount       int      `json:"followers_count"`
	FriendsCount         int      `json:"friends_count"`
	ID                   int64    `json:"id_str,string"`
	ListedCount          int      `json:"listed_count"`
	Name                 string   `json:"name"`
	Location             string   `json:"location"`
	PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"` // Dunno how to type-convert an array
	ProfileBannerURL     string   `json:"profile_banner_url"`
	ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
	Protected            bool     `json:"protected"`
	ScreenName           string   `json:"screen_name"`
	StatusesCount        int      `json:"statuses_count"`
	Verified             bool     `json:"verified"`
	IsBanned             bool
	DoesntExist          bool
}

// Turn an APIUser, as returned from the scraper, into a properly structured User object
func ParseSingleUser(apiUser APIUser) (ret User, err error) {
	if apiUser.DoesntExist {
		// User may have been deleted, or there was a typo.  There's no data to parse
		if apiUser.ScreenName == "" {
			panic("ScreenName is empty!")
		}
		ret = GetUnknownUserWithHandle(UserHandle(apiUser.ScreenName))
		return
	}
	ret.ID = UserID(apiUser.ID)
	ret.Handle = UserHandle(apiUser.ScreenName)
	if apiUser.IsBanned {
		// Banned users won't have any further info, so just return here
		ret.IsBanned = true
		return
	}
	ret.DisplayName = apiUser.Name
	ret.Bio = apiUser.Description
	ret.FollowingCount = apiUser.FriendsCount
	ret.FollowersCount = apiUser.FollowersCount
	ret.Location = apiUser.Location
	if len(apiUser.Entities.URL.Urls) > 0 {
		ret.Website = apiUser.Entities.URL.Urls[0].ExpandedURL
	}
	ret.JoinDate, err = TimestampFromString(apiUser.CreatedAt)
	if err != nil {
		err = fmt.Errorf("Error parsing time on user ID %d: %w", ret.ID, err)
		return
	}
	ret.IsPrivate = apiUser.Protected
	ret.IsVerified = apiUser.Verified
	ret.ProfileImageUrl = apiUser.ProfileImageURLHTTPS

	if regexp.MustCompile(`_normal\.\w{2,4}`).MatchString(ret.ProfileImageUrl) {
		ret.ProfileImageUrl = strings.ReplaceAll(ret.ProfileImageUrl, "_normal.", ".")
	}
	ret.BannerImageUrl = apiUser.ProfileBannerURL

	ret.ProfileImageLocalPath = ret.compute_profile_image_local_path()
	ret.BannerImageLocalPath = ret.compute_banner_image_local_path()

	if len(apiUser.PinnedTweetIdsStr) > 0 {
		ret.PinnedTweetID = TweetID(idstr_to_int(apiUser.PinnedTweetIdsStr[0]))
	}
	return
}

type APINotification struct {
	ID          string `json:"id"`
	TimestampMs int64  `json:"timestampMs,string"`
	Message     struct {
		Text     string `json:"text"`
		Entities []struct {
			FromIndex int `json:"fromIndex"`
			ToIndex   int `json:"toIndex"`
			Ref       struct {
				User struct {
					ID int `json:"id,string"`
				} `json:"user"`
			} `json:"ref"`
		} `json:"entities"`
	} `json:"message"`
	Template struct {
		AggregateUserActionsV1 struct {
			TargetObjects []struct {
				Tweet struct {
					ID int `json:"id,string"`
				} `json:"tweet"`
			} `json:"targetObjects"`
			FromUsers []struct {
				User struct {
					ID int `json:"id,string"`
				} `json:"user"`
			} `json:"fromUsers"`
		} `json:"aggregateUserActionsV1"`
	} `json:"template"`
}

type UserResponse struct {
	Data struct {
		User struct {
			Result struct {
				MetaTypename       string  `json:"__typename"`
				ID                 int64   `json:"rest_id,string"`
				Legacy             APIUser `json:"legacy"`
				IsBlueVerified     bool    `json:"is_blue_verified"`
				UnavailableMessage struct {
					Text string `json:"text"`
				} `json:"unavailable_message"`
				Reason string `json:"reason"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Name    string `json:"name"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

func (u UserResponse) ConvertToAPIUser() (APIUser, error) {
	if u.Data.User.Result.MetaTypename == "" {
		// Completely empty response (user not found)
		return APIUser{}, ErrDoesntExist
	}

	ret := u.Data.User.Result.Legacy
	ret.ID = u.Data.User.Result.ID
	ret.Verified = u.Data.User.Result.IsBlueVerified

	// Banned users
	for _, api_error := range u.Errors {
		if api_error.Message == "Authorization: User has been suspended. (63)" {
			ret.IsBanned = true
		} else if api_error.Name == "NotFoundError" {
			ret.DoesntExist = true
		} else {
			panic(fmt.Errorf("Unknown api error %q:\n  %w", api_error.Message, EXTERNAL_API_ERROR))
		}
	}

	// Banned users, new version
	if u.Data.User.Result.Reason == "Suspended" {
		ret.IsBanned = true
	}

	// Deleted users
	if ret.ID == 0 && ret.ScreenName == "" && u.Data.User.Result.Reason != "Suspended" {
		ret.DoesntExist = true
	}

	return ret, nil
}

type Entry struct {
	EntryID   string `json:"entryId"`
	SortIndex int64  `json:"sortIndex,string"`
	Content   struct {
		Item struct {
			Content struct {
				Tombstone struct {
					TombstoneInfo struct {
						RichText struct {
							Text string `json:"text"`
						} `json:"richText"`
					} `json:"tombstoneInfo"`
				} `json:"tombstone"`
				Tweet struct {
					ID int64 `json:"id,string"`
				} `json:"tweet"`
				Notification struct {
					ID           string     `json:"id"`
					FromUsers    Int64Slice `json:"fromUsers"`
					TargetTweets Int64Slice `json:"targetTweets"`
				} `json:"notification"`
			} `json:"content"`
			ClientEventInfo struct {
				Element string `json:"element"`
			} `json:"clientEventInfo"`
		} `json:"item"`
		Operation struct {
			Cursor struct {
				Value string `json:"value"`
			} `json:"cursor"`
		} `json:"operation"`
	} `json:"content"`
}

func (e Entry) GetTombstoneText() string {
	return e.Content.Item.Content.Tombstone.TombstoneInfo.RichText.Text
}

type SortableEntries []Entry

func (e SortableEntries) Len() int           { return len(e) }
func (e SortableEntries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e SortableEntries) Less(i, j int) bool { return e[i].SortIndex > e[j].SortIndex }

type TweetResponse struct {
	GlobalObjects struct {
		Tweets        map[string]APITweet        `json:"tweets"`
		Users         map[string]APIUser         `json:"users"`
		Notifications map[string]APINotification `json:"notifications"`
	} `json:"globalObjects"`
	Timeline struct {
		Instructions []struct {
			AddEntries struct {
				Entries SortableEntries `json:"entries"`
			} `json:"addEntries"`
			ReplaceEntry struct {
				Entry Entry
			} `json:"replaceEntry"`
			MarkEntriesUnreadGreaterThanSortIndex struct {
				SortIndex int64 `json:"sortIndex,string"`
			} `json:"markEntriesUnreadGreaterThanSortIndex"`
		} `json:"instructions"`
	} `json:"timeline"`
}

var tombstone_types = map[string]string{
	"This Tweet was deleted by the Tweet author. Learn more":                                                   "deleted",
	"This Tweet is from a suspended account. Learn more":                                                       "suspended",
	"You’re unable to view this Tweet because this account owner limits who can view their Tweets. Learn more": "hidden",
	"This Tweet is unavailable. Learn more":                                                                    "unavailable",
	"This Tweet violated the Twitter Rules. Learn more":                                                        "violated",
	"This Tweet is from an account that no longer exists. Learn more":                                          "no longer exists",
	"Age-restricted adult content. This content might not be appropriate for people under 18 years old. To view this media, " +
		"you’ll need to log in to Twitter. Learn more": "age-restricted",

	// New versions that use "Post" instead of "Tweet" and "X" instead of "Twitter"
	"This Post was deleted by the Post author. Learn more":                                                   "deleted",
	"This Post is from a suspended account. Learn more":                                                      "suspended",
	"You’re unable to view this Post because this account owner limits who can view their Posts. Learn more": "hidden",
	"This Post is unavailable. Learn more":                                                                   "unavailable",
	"This Post violated the X Rules. Learn more":                                                             "violated",
	"This Post is from an account that no longer exists. Learn more":                                         "no longer exists",
}

/**
 * Insert tweets into GlobalObjects for each tombstone.  Returns a list of users that need to
 * be fetched for tombstones.
 */
func (t *TweetResponse) HandleTombstones() []UserHandle {
	ret := []UserHandle{}

	// Handle tombstones in quote-tweets
	for _, api_tweet := range t.GlobalObjects.Tweets {
		// Ignore if tweet doesn't have a quoted tweet
		if api_tweet.QuotedStatusIDStr == "" {
			continue
		}
		// Ignore if quoted tweet is in the Global Objects (i.e., not a tombstone)
		if _, ok := t.GlobalObjects.Tweets[api_tweet.QuotedStatusIDStr]; ok {
			continue
		}

		user_handle, err := ParseHandleFromTweetUrl(api_tweet.QuotedStatusPermalink.ExpandedURL)
		if err != nil {
			panic(err)
		}

		var tombstoned_tweet APITweet
		tombstoned_tweet.ID = int64(int_or_panic(api_tweet.QuotedStatusIDStr))
		tombstoned_tweet.UserHandle = string(user_handle)
		tombstoned_tweet.TombstoneText = "unavailable"

		ret = append(ret, user_handle)
		fmt.Printf("Adding quoted tombstoned tweet: TweetID %d, handle %q\n", tombstoned_tweet.ID, tombstoned_tweet.UserHandle)

		t.GlobalObjects.Tweets[api_tweet.QuotedStatusIDStr] = tombstoned_tweet
	}

	// Handle tombstones in the conversation flow
	entries := t.Timeline.Instructions[0].AddEntries.Entries
	sort.Sort(entries)
	for i, entry := range entries {
		if entry.GetTombstoneText() != "" {
			// Try to reconstruct the tombstone tweet
			var tombstoned_tweet APITweet
			tombstoned_tweet.ID = int64(i) // Set a default to prevent clobbering other tombstones
			if i+1 < len(entries) && entries[i+1].Content.Item.Content.Tweet.ID != 0 {
				next_tweet_id := entries[i+1].Content.Item.Content.Tweet.ID
				api_tweet, ok := t.GlobalObjects.Tweets[fmt.Sprint(next_tweet_id)]
				if !ok {
					panic("Weird situation!")
				}
				tombstoned_tweet.ID = api_tweet.InReplyToStatusID
				tombstoned_tweet.UserID = api_tweet.InReplyToUserID
				ret = append(ret, UserHandle(api_tweet.InReplyToScreenName))
			}
			if i-1 >= 0 && entries[i-1].Content.Item.Content.Tweet.ID != 0 {
				prev_tweet_id := entries[i-1].Content.Item.Content.Tweet.ID
				_, ok := t.GlobalObjects.Tweets[fmt.Sprint(prev_tweet_id)]
				if !ok {
					panic("Weird situation 2!")
				}
				tombstoned_tweet.InReplyToStatusID = prev_tweet_id
			}

			short_text, ok := tombstone_types[entry.GetTombstoneText()]
			if !ok {
				panic(fmt.Errorf("Unknown tombstone text %q:\n  %w", entry.GetTombstoneText(), EXTERNAL_API_ERROR))
			}
			tombstoned_tweet.TombstoneText = short_text

			// Add the tombstoned tweet to GlobalObjects
			t.GlobalObjects.Tweets[fmt.Sprint(tombstoned_tweet.ID)] = tombstoned_tweet
		}
	}

	return ret
}

func (t *TweetResponse) GetCursor() string {
	// TODO: is this function used anywhere other than Notifications?
	for _, instr := range t.Timeline.Instructions {
		if len(instr.AddEntries.Entries) > 0 {
			last_entry := instr.AddEntries.Entries[len(instr.AddEntries.Entries)-1]
			if strings.Contains(last_entry.EntryID, "cursor") {
				return last_entry.Content.Operation.Cursor.Value
			}
		}
	}

	// Next, try the other format ("replaceEntry")
	instructions := t.Timeline.Instructions
	last_replace_entry := instructions[len(instructions)-1].ReplaceEntry.Entry
	if strings.Contains(last_replace_entry.EntryID, "cursor") {
		return last_replace_entry.Content.Operation.Cursor.Value
	}
	return ""
}

func (t *TweetResponse) GetCursorTop() string {
	for _, instr := range t.Timeline.Instructions {
		for _, entry := range instr.AddEntries.Entries {
			if strings.Contains(entry.EntryID, "cursor-top") {
				return entry.Content.Operation.Cursor.Value
			}
		}
	}
	return ""
}

/**
 * Test for one case of end-of-feed.  Cursor increments on each request for some reason, but
 * there's no new content.  This seems to happen when there's a pinned tweet.
 *
 * In this case, we look for an "entries" object that has only cursors in it, and no tweets.
 */
func (t *TweetResponse) IsEndOfFeed() bool {
	for _, instr := range t.Timeline.Instructions {
		entries := instr.AddEntries.Entries
		if len(entries) == 0 {
			continue // Not the main instruction
		}
		if len(entries) > 2 {
			return false
		}
		for _, e := range entries {
			if !strings.Contains(e.EntryID, "cursor") {
				return false
			}
		}
	}
	return true
}

func (t *TweetResponse) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()

	for _, single_tweet := range t.GlobalObjects.Tweets {
		trove, err := single_tweet.ToTweetTrove()
		if err != nil {
			return ret, err
		}
		ret.MergeWith(trove)
	}

	for _, user := range t.GlobalObjects.Users {
		new_user, err := ParseSingleUser(user)
		if err != nil {
			return ret, err
		}
		ret.Users[new_user.ID] = new_user
	}
	for _, n := range t.GlobalObjects.Notifications {
		new_notification := ParseSingleNotification(n)
		ret.Notifications[new_notification.ID] = new_notification
	}
	return ret, nil
}

func idstr_to_int(s string) int64 {
	return int64(int_or_panic(s))
}

func int_or_panic(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}
