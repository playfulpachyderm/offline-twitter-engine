package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

var ErrorIsTombstone = errors.New("tweet is a tombstone")

type CardValue struct {
	Type        string `json:"type"`
	StringValue string `json:"string_value"`
	ImageValue  struct {
		AltText string `json:"alt"`
		Height  int    `json:"height"`
		Width   int    `json:"width"`
		Url     string `json:"url"`
	} `json:"image_value"`
	UserValue struct {
		ID int64 `json:"id_str,string"`
	} `json:"user_value"`
	BooleanValue bool `json:"boolean_value"`
}

type APIV2Card struct {
	Legacy struct {
		BindingValues []struct {
			Key   string    `json:"key"`
			Value CardValue `json:"value"`
		} `json:"binding_values"`
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"legacy"`
}

func (card APIV2Card) ParseAsUrl() Url {
	values := make(map[string]CardValue)
	for _, obj := range card.Legacy.BindingValues {
		values[obj.Key] = obj.Value
	}

	ret := Url{}
	ret.HasCard = true

	ret.ShortText = card.Legacy.Url
	ret.Domain = values["domain"].StringValue
	ret.Title = values["title"].StringValue
	ret.Description = values["description"].StringValue
	ret.IsContentDownloaded = false
	ret.CreatorID = UserID(values["creator"].UserValue.ID)
	ret.SiteID = UserID(values["site"].UserValue.ID)

	var thumbnail_url string
	if card.Legacy.Name == "summary_large_image" || card.Legacy.Name == "summary" {
		thumbnail_url = values["thumbnail_image_large"].ImageValue.Url
	} else if card.Legacy.Name == "player" {
		thumbnail_url = values["player_image_large"].ImageValue.Url
	} else {
		panic("TODO unknown card type")
	}

	if thumbnail_url != "" {
		ret.HasThumbnail = true
		ret.ThumbnailRemoteUrl = thumbnail_url
		ret.ThumbnailLocalPath = get_thumbnail_local_path(thumbnail_url)
		ret.ThumbnailWidth = values["thumbnail_image_large"].ImageValue.Width
		ret.ThumbnailHeight = values["thumbnail_image_large"].ImageValue.Height
	}
	return ret
}
func (card APIV2Card) ParseAsPoll() Poll {
	values := make(map[string]CardValue)
	for _, obj := range card.Legacy.BindingValues {
		values[obj.Key] = obj.Value
	}

	card_url, err := url.Parse(card.Legacy.Url)
	if err != nil {
		panic(err)
	}
	id := int_or_panic(card_url.Hostname())

	ret := Poll{}
	ret.ID = PollID(id)
	ret.NumChoices = parse_num_choices(card.Legacy.Name)
	ret.VotingDuration = int_or_panic(values["duration_minutes"].StringValue) * 60
	ret.VotingEndsAt, err = TimestampFromString(values["end_datetime_utc"].StringValue)
	if err != nil {
		panic(err)
	}
	ret.LastUpdatedAt, err = TimestampFromString(values["last_updated_datetime_utc"].StringValue)
	if err != nil {
		panic(err)
	}

	ret.Choice1 = values["choice1_label"].StringValue
	ret.Choice1_Votes = int_or_panic(values["choice1_count"].StringValue)
	ret.Choice2 = values["choice2_label"].StringValue
	ret.Choice2_Votes = int_or_panic(values["choice2_count"].StringValue)

	if ret.NumChoices > 2 {
		ret.Choice3 = values["choice3_label"].StringValue
		ret.Choice3_Votes = int_or_panic(values["choice3_count"].StringValue)
	}
	if ret.NumChoices > 3 {
		ret.Choice4 = values["choice4_label"].StringValue
		ret.Choice4_Votes = int_or_panic(values["choice4_count"].StringValue)
	}
	return ret
}
func (card APIV2Card) ParseAsSpace() Space {
	values := make(map[string]CardValue)
	for _, obj := range card.Legacy.BindingValues {
		values[obj.Key] = obj.Value
	}
	ret := Space{}
	ret.ID = SpaceID(values["id"].StringValue)
	ret.ShortUrl = values["card_url"].StringValue

	return ret
}

type APIV2UserResult struct {
	UserResults struct {
		Result struct {
			ID     int64   `json:"rest_id,string"`
			Legacy APIUser `json:"legacy"`
		} `json:"result"`
	} `json:"user_results"`
}

func (u APIV2UserResult) ToUser() User {
	user, err := ParseSingleUser(u.UserResults.Result.Legacy)
	if err != nil {
		panic(err)
	}
	user.ID = UserID(u.UserResults.Result.ID)
	return user
}

type _Result struct {
	ID        int64      `json:"rest_id,string"`
	Legacy    APIV2Tweet `json:"legacy"`
	Tombstone *struct {
		Text struct {
			Text string `json:"text"`
		} `json:"text"`
	} `json:"tombstone"`
	Core               *APIV2UserResult `json:"core"`
	Card               APIV2Card        `json:"card"`
	QuotedStatusResult *APIV2Result     `json:"quoted_status_result"`
	NoteTweet          struct {
		IsExpandable     bool `json:"is_expandable"`
		NoteTweetResults struct {
			Result struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"result"`
		} `json:"note_tweet_results"`
	} `json:"note_tweet"`
}

type APIV2Result struct {
	Result struct {
		_Result
		Tweet _Result `json:"tweet"`
	} `json:"result"`
}

func (api_result APIV2Result) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()

	// Start by checking if this is a null entry in a feed
	if api_result.Result.Tombstone != nil {
		// Returning an error indicates the parent (APIV2Entry) has to parse it as a tombstone.
		// The tweet ID isn't available to the APIV2Result, but it is to the APIV2Entry.
		return ret, ErrorIsTombstone
	}

	if api_result.Result.Legacy.ID == 0 && api_result.Result.Tweet.Legacy.ID != 0 {
		// If the tweet has "__typename" of "TweetWithVisibilityResults", it uses a new structure with
		// a "tweet" field with the regular data, alongside a "tweetInterstitial" field which is ignored
		// for now.
		log.Debug("Using Inner Tweet")
		api_result.Result._Result = api_result.Result.Tweet
	}

	// Handle expandable tweets
	if api_result.Result.NoteTweet.IsExpandable {
		api_result.Result.Legacy.FullText = api_result.Result.NoteTweet.NoteTweetResults.Result.Text
		api_result.Result.Legacy.DisplayTextRange = []int{} // Override the "display text"
		api_result.Result.Legacy.IsExpandable = true
	}

	// Process the tweet itself
	main_tweet_trove := api_result.Result.Legacy.ToTweetTrove()
	ret.MergeWith(main_tweet_trove)

	// Parse the User info
	if api_result.Result.Core != nil {
		// `Core` is nil for tombstones because they don't carry user info.  Nothing to do here
		main_user := api_result.Result.Core.ToUser()
		ret.Users[main_user.ID] = main_user
	}

	// Handle quoted tweet
	if api_result.Result.QuotedStatusResult != nil {
		quoted_api_result := api_result.Result.QuotedStatusResult
		quoted_trove, err := quoted_api_result.ToTweetTrove()

		// Quoted tombstones can be handled here since we already have the ID and user handle
		if errors.Is(err, ErrorIsTombstone) {
			tombstoned_tweet := quoted_api_result.Result.Legacy.APITweet

			// Capture the tombstone text
			var is_ok bool
			tombstoned_tweet.TombstoneText, is_ok = tombstone_types[quoted_api_result.Result.Tombstone.Text.Text]
			if !is_ok {
				panic(fmt.Errorf("Unknown tombstone text %q:\n  %w", quoted_api_result.Result.Tombstone.Text.Text, EXTERNAL_API_ERROR))
			}

			// Capture the tombstone ID
			tombstoned_tweet.ID = int64(int_or_panic(api_result.Result.Legacy.APITweet.QuotedStatusIDStr))

			// Capture the tombstone's user handle
			handle, err := ParseHandleFromTweetUrl(api_result.Result.Legacy.APITweet.QuotedStatusPermalink.ExpandedURL)
			if err != nil {
				panic(err)
			}
			tombstoned_tweet.UserHandle = string(handle)

			// Parse the tombstone into a Tweet and add it to the trove
			parsed_tombstone_tweet, err := ParseSingleTweet(tombstoned_tweet)
			if err != nil {
				panic(err)
			}
			ret.Tweets[parsed_tombstone_tweet.ID] = parsed_tombstone_tweet

			// Add the user as a tombstoned user to be fetched later
			ret.TombstoneUsers = append(ret.TombstoneUsers, handle)
		} else if err != nil {
			panic(err)
		}

		ret.MergeWith(quoted_trove)
	}

	// Handle URL cards.
	// This should be done in APIV2Tweet (not APIV2Result), but due to the terrible API response structuring (the Card
	// should be nested under the APIV2Tweet, but it isn't), it goes here.
	if api_result.Result.Legacy.RetweetedStatusResult == nil {
		// We have to filter out retweets.  For some reason, retweets have a copy of the card in both the retweeting
		// and the retweeted TweetResults; it should only be parsed for the real Tweet, not the Retweet
		main_tweet, is_ok := ret.Tweets[TweetID(api_result.Result.Legacy.ID)]
		if !is_ok {
			panic(fmt.Errorf("Tweet trove didn't contain its own tweet with ID %d:\n  %w", api_result.Result.Legacy.ID, EXTERNAL_API_ERROR))
		}
		if api_result.Result.Card.Legacy.Name == "summary_large_image" || api_result.Result.Card.Legacy.Name == "player" {
			url := api_result.Result.Card.ParseAsUrl()

			url.TweetID = main_tweet.ID
			found := false
			for i := range main_tweet.Urls {
				if main_tweet.Urls[i].ShortText != url.ShortText {
					continue
				}
				found = true
				url.Text = main_tweet.Urls[i].Text // Copy the expanded URL over, since the card doesn't have it in the new API
				main_tweet.Urls[i] = url
			}
			if !found {
				panic(fmt.Errorf("Couldn't find the url in tweet ID %d:\n  %w", api_result.Result.Legacy.ID, EXTERNAL_API_ERROR))
			}
		} else if strings.Index(api_result.Result.Card.Legacy.Name, "poll") == 0 {
			// Process polls
			poll := api_result.Result.Card.ParseAsPoll()
			poll.TweetID = main_tweet.ID
			main_tweet.Polls = []Poll{poll}
			ret.Tweets[main_tweet.ID] = main_tweet
		} else if api_result.Result.Card.Legacy.Name == "3691233323:audiospace" {
			space := api_result.Result.Card.ParseAsSpace()
			// Attach it to the Tweet that linked it
			main_tweet.SpaceID = space.ID
			// Put it in the trove; avoid clobbering
			if existing_space, is_ok := ret.Spaces[space.ID]; !is_ok || !existing_space.IsDetailsFetched {
				ret.Spaces[space.ID] = space
			}

			// main_tweet.Spaces = []Space{space}

			// Remove it from the Urls
			for i, url := range main_tweet.Urls {
				if url.ShortText == space.ShortUrl {
					main_tweet.Urls = append(main_tweet.Urls[:i], main_tweet.Urls[i+1:]...)
					break
				}
			}

			ret.Tweets[main_tweet.ID] = main_tweet
		}
	}

	return ret, nil
}

type APIV2Tweet struct {
	// For some reason, retweets are nested *inside* the Legacy tweet, whereas
	// quoted-tweets are next to it, as their own tweet
	RetweetedStatusResult *APIV2Result `json:"retweeted_status_result"`
	APITweet
}

func (api_v2_tweet APIV2Tweet) ToTweetTrove() TweetTrove {
	ret := NewTweetTrove()

	// If there's a retweet, we ignore the main tweet except for posted_at and retweeting UserID
	if api_v2_tweet.RetweetedStatusResult != nil {
		orig_tweet_trove, err := api_v2_tweet.RetweetedStatusResult.ToTweetTrove()
		if err != nil {
			panic(err)
		}
		ret.MergeWith(orig_tweet_trove)

		retweet := Retweet{}

		retweet.RetweetID = TweetID(api_v2_tweet.ID)
		if api_v2_tweet.RetweetedStatusResult.Result.Legacy.ID == 0 && api_v2_tweet.RetweetedStatusResult.Result.Tweet.Legacy.ID != 0 {
			// Tweet is a "TweetWithVisibilityResults" (See above comment for more).
			retweet.TweetID = TweetID(api_v2_tweet.RetweetedStatusResult.Result.Tweet.ID)
		} else {
			retweet.TweetID = TweetID(api_v2_tweet.RetweetedStatusResult.Result.ID)
		}

		retweet.RetweetedByID = UserID(api_v2_tweet.APITweet.UserID)
		retweet.RetweetedAt, err = TimestampFromString(api_v2_tweet.APITweet.CreatedAt)
		if err != nil {
			fmt.Printf("%v\n", api_v2_tweet)
			panic(err)
		}
		ret.Retweets[retweet.RetweetID] = retweet
	} else {
		main_tweet, err := ParseSingleTweet(api_v2_tweet.APITweet)
		if err != nil {
			panic(err)
		}
		ret.Tweets[main_tweet.ID] = main_tweet
	}

	return ret
}

type ItemContent struct {
	ItemType     string      `json:"itemType"`
	TweetResults APIV2Result `json:"tweet_results"`

	// Cursors (conversation view format)
	CursorType string `json:"cursorType"`
	Value      string `json:"value"`
}

// Wraps InnerAPIV2Entry to implement `json.Unmarshal`.  Does the normal unmarshal but also saves the original JSON.
type APIV2Entry struct {
	InnerAPIV2Entry
	OriginalJSON string
}
type InnerAPIV2Entry struct {
	EntryID   string `json:"entryId"`
	SortIndex int64  `json:"sortIndex,string"`
	Content   struct {
		ItemContent ItemContent `json:"itemContent"`

		Items []struct {
			EntryId     string
			Dispensable bool
			Item        struct {
				ItemContent ItemContent `json:"itemContent"`
			}
		}

		// Cursors (user feed format)
		EntryType  string `json:"entryType"`
		Value      string `json:"value"`
		CursorType string `json:"cursorType"`
	} `json:"content"`
}

func (e *APIV2Entry) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &e.InnerAPIV2Entry)
	if err != nil {
		return fmt.Errorf("Error parsing json APIV2Entry:\n  %w", err)
	}
	e.OriginalJSON = string(data)
	return nil
}

// Parse the entry's "entryId" field to infer the ID of the tweet it contains.
// The returned TweetID is only valid if the entryID type is "tweet" (i.e., the entryID is like "tweet-12345").
// Return the entry type and the TweetID.
func (e *APIV2Entry) ParseID() (string, TweetID) {
	parts := strings.Split(e.EntryID, "-")
	if len(parts) < 2 {
		panic(fmt.Sprintf("Entry id (%s) has %d parts!", e.EntryID, len(parts)))
	}
	return strings.Join(parts[0:len(parts)-1], "-"), TweetID(int_or_panic(parts[len(parts)-1]))
}

func (e APIV2Entry) ToTweetTrove() TweetTrove {
	defer func() {
		if obj := recover(); obj != nil {
			log.Warn(fmt.Sprintf("Panic while decoding entry: %s\n", e.OriginalJSON))
			panic(obj)
		}
	}()
	if e.Content.EntryType == "TimelineTimelineCursor" || e.Content.ItemContent.ItemType == "TimelineTimelineCursor" {
		// Ignore cursor entries.
		// - e.Content.EntryType -> User Feed itself
		// - e.Content.ItemContent.ItemType -> conversation thread in a user feed
		return NewTweetTrove()
	} else if e.Content.EntryType == "TimelineTimelineModule" {
		ret := NewTweetTrove()

		switch strings.Split(e.EntryID, "-")[0] {
		case "homeConversation", "conversationthread":
			// Process it.
			// - "homeConversation": conversation thread on a user feed
			// - "conversationthread": conversation thread in the replies under a TweetDetail view
			for _, item := range e.Content.Items {
				if item.Item.ItemContent.ItemType == "TimelineTimelineCursor" {
					// "Show More" replies button in a thread on Tweet Detail page
					continue
				}
				trove, err := item.Item.ItemContent.TweetResults.ToTweetTrove()
				if err != nil {
					panic(err)
				}
				ret.MergeWith(trove)
			}

		case "whoToFollow", "TopicsModule", "tweetdetailrelatedtweets":
			// Ignore "Who to follow", "Topics" and "Related Tweets" modules.
			// TODO: maybe we can capture these eventually
			log.Debug(fmt.Sprintf("Skipping %s entry", e.EntryID))

		default:
			log.Warn("TimelineTimelineModule with unknown EntryID: " + e.EntryID)
		}

		return ret
	} else if e.Content.EntryType == "TimelineTimelineItem" {
		ret, err := e.Content.ItemContent.TweetResults.ToTweetTrove()

		if errors.Is(err, ErrorIsTombstone) {
			// Handle tombstones
			ret = NewTweetTrove()                                                         // clear the result just in case
			tombstoned_tweet := e.Content.ItemContent.TweetResults.Result.Legacy.APITweet // Will be empty to start

			// Capture the tombstone text
			var is_ok bool
			tombstoned_tweet.TombstoneText, is_ok = tombstone_types[e.Content.ItemContent.TweetResults.Result.Tombstone.Text.Text]
			if !is_ok {
				panic(fmt.Errorf(
					"Unknown tombstone text %q:\n  %w",
					e.Content.ItemContent.TweetResults.Result.Tombstone.Text.Text,
					EXTERNAL_API_ERROR,
				))
			}

			// Capture the tombstone ID
			_, tombstoned_tweet_id := e.ParseID()
			tombstoned_tweet.ID = int64(tombstoned_tweet_id)

			// Parse the tombstone into a Tweet and add it to the trove
			parsed_tombstone_tweet, err := ParseSingleTweet(tombstoned_tweet)
			if err != nil {
				panic(err)
			}
			ret.Tweets[parsed_tombstone_tweet.ID] = parsed_tombstone_tweet
		}
		return ret
	}
	panic("Unknown EntryType: " + e.Content.EntryType)
}

type APIV2Instruction struct {
	Type    string       `json:"type"`
	Entries []APIV2Entry `json:"entries"`
}

type APIV2Response struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []APIV2Instruction `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline"`
			} `json:"result"`
		} `json:"user"`
		ThreadedConversationWithInjectionsV2 struct {
			Instructions []APIV2Instruction `json:"instructions"`
		} `json:"threaded_conversation_with_injections_v2"`
	} `json:"data"`
}

func (api_response APIV2Response) GetMainInstruction() *APIV2Instruction {
	instructions := api_response.Data.User.Result.Timeline.Timeline.Instructions
	for i := range instructions {
		if instructions[i].Type == "TimelineAddEntries" {
			return &instructions[i]
		}
	}
	instructions = api_response.Data.ThreadedConversationWithInjectionsV2.Instructions
	for i := range instructions {
		if instructions[i].Type == "TimelineAddEntries" {
			return &instructions[i]
		}
	}
	panic("No 'TimelineAddEntries' found")
}

func (api_response APIV2Response) GetCursorBottom() string {
	for _, entry := range api_response.GetMainInstruction().Entries {
		// For a user feed:
		if entry.Content.CursorType == "Bottom" {
			return entry.Content.Value
		}

		// For a Tweet Detail page:
		if entry.Content.ItemContent.CursorType == "Bottom" {
			return entry.Content.ItemContent.Value
		}
	}
	return ""
}

/**
 * Returns `true` if there's no non-cursor entries in this response, false otherwise
 */
func (api_response APIV2Response) IsEmpty() bool {
	for _, e := range api_response.GetMainInstruction().Entries {
		if !strings.Contains(e.EntryID, "cursor") {
			return false
		}
	}
	return true
}

/**
 * Parse the collected API response and turn it into a TweetTrove
 */
func (api_response APIV2Response) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()

	// Parse all of the entries
	for _, entry := range api_response.GetMainInstruction().Entries { // TODO: the second Instruction is the pinned tweet
		main_trove := entry.ToTweetTrove()
		ret.MergeWith(main_trove)
	}

	// Add in any tombstoned user handles and IDs if possible, by reading from the replies
	for _, tweet := range ret.Tweets {
		// Skip if it's not a reply (nothing to add)
		if tweet.InReplyToID == 0 {
			continue
		}

		// Skip if the replied tweet isn't in the result set (e.g., the reply is a quoted tweet)
		replied_tweet, is_ok := ret.Tweets[tweet.InReplyToID]
		if !is_ok {
			continue
		}

		// Skip if the replied tweet isn't a stub (it's already filled out)
		if !replied_tweet.IsStub {
			continue
		}

		if replied_tweet.ID == 0 {
			// Not sure if this can happen.  It should get filled out by parsing the entry ID.
			// Use a panic to detect if it does so we can analyse
			// TODO: make a better system to capture "discovery panics" that doesn't involve panicking
			panic(fmt.Sprintf("Tombstoned tweet has no ID (should be %d)", tweet.InReplyToID))
		}

		if replied_tweet.UserID == 0 {
			replied_tweet.UserID = tweet.in_reply_to_user_id
			if replied_tweet.UserID == 0 { // Still??
				log.Warn(fmt.Sprintf("Still couldn't find user for replied tweet %d", tweet.InReplyToID))
				continue
			}
		} // replied_tweet.UserID should now be a real UserID

		existing_user, is_ok := ret.Users[replied_tweet.UserID]
		if !is_ok {
			existing_user = User{ID: replied_tweet.UserID}
		}
		if existing_user.Handle == "" {
			existing_user.Handle = tweet.in_reply_to_user_handle
		}
		ret.Users[replied_tweet.UserID] = existing_user
		// TODO: add to ret.TombstonedUsers?

		ret.Tweets[replied_tweet.ID] = replied_tweet
	}

	return ret, nil // TODO: This doesn't need to return an error, it's always nil
}

func get_graphql_user_timeline_url(user_id UserID, cursor string) string {
	if cursor != "" {
		return "https://twitter.com/i/api/graphql/CwLU7qTfeu0doqhSr6tW4A/UserTweetsAndReplies?variables=%7B%22userId%22%3A%22" + fmt.Sprint(user_id) + "%22%2C%22count%22%3A40%2C%22cursor%22%3A%22" + url.QueryEscape(cursor) + "%22%2C%22includePromotedContent%22%3Atrue%2C%22withCommunity%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withBirdwatchPivots%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%2C%22__fs_interactive_text%22%3Afalse%2C%22__fs_responsive_web_uc_gql_enabled%22%3Afalse%2C%22__fs_dont_mention_me_view_api_enabled%22%3Afalse%7D" //nolint:lll  // It's a URL, come on
	}
	return "https://twitter.com/i/api/graphql/CwLU7qTfeu0doqhSr6tW4A/UserTweetsAndReplies?variables=%7B%22userId%22%3A%22" + fmt.Sprint(user_id) + "%22%2C%22count%22%3A40%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withBirdwatchPivots%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%2C%22__fs_interactive_text%22%3Afalse%2C%22__fs_dont_mention_me_view_api_enabled%22%3Afalse%7D" //nolint:lll  // It's a URL, come on
}

/**
 * Get a User feed using the new GraphQL twitter api
 */
func (api *API) GetGraphqlFeedFor(user_id UserID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(get_graphql_user_timeline_url(user_id, cursor))
	if err != nil {
		panic(err)
	}

	var response APIV2Response
	err = api.do_http(url.String(), cursor, &response)

	return response, err
}

func (api API) GetLikesFor(user_id UserID, cursor string) (APIV2Response, error) {
	var response APIV2Response
	err := api.do_http("https://twitter.com/i/api/graphql/2Z6LYO4UTM4BnWjaNCod6g/Likes?variables=%7B%22userId%22%3A%22"+fmt.Sprint(user_id)+"%22%2C%22count%22%3A20%2C%22includePromotedContent%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withClientEventToken%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%7D&features=%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D", cursor, &response) //nolint:lll  // It's a URL, come on
	return response, err
}

/**
 * Resend the request to get more tweets if necessary
 *
 * args:
 * - user_id: the user's UserID
 * - response: an "out" parameter; the APIV2Response that tweets, RTs and users will be appended to
 * - min_tweets: the desired minimum amount of tweets to get
 */
func (api *API) GetMoreTweetsFromGraphqlFeed(user_id UserID, response *APIV2Response, min_tweets int) error {
	// TODO user-feed-infinite-fetch: what if you reach the end of the user's timeline?  Might loop
	// forever getting no new tweets
	last_response := response
	for last_response.GetCursorBottom() != "" && len(response.GetMainInstruction().Entries) < min_tweets {
		fresh_response, err := api.GetGraphqlFeedFor(user_id, last_response.GetCursorBottom())
		if err != nil {
			return err
		}

		if fresh_response.GetCursorBottom() == last_response.GetCursorBottom() && len(fresh_response.GetMainInstruction().Entries) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			return END_OF_FEED
		}
		if fresh_response.IsEmpty() {
			// Response has a pinned tweet, but no other content: end of feed has been reached
			return END_OF_FEED // TODO: check that there actually is a pinned tweet and the request didn't just fail lol
		}

		last_response = &fresh_response

		// Copy over the entries
		response.GetMainInstruction().Entries = append(
			response.GetMainInstruction().Entries,
			last_response.GetMainInstruction().Entries...)

		fmt.Printf("Have %d entries so far\n", len(response.GetMainInstruction().Entries))
	}
	return nil
}
