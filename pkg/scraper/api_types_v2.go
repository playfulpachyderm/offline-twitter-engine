package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

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

type Int64Slice []int64

func (s *Int64Slice) UnmarshalJSON(data []byte) error {
	var result []string

	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}

	for _, str := range result {
		num, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			panic(err)
		}
		*s = append(*s, num)
	}
	return nil
}

type Tombstone struct {
	Text struct {
		Text string `json:"text"`
	} `json:"text"`
}
type _Result struct {
	ID                 int64            `json:"rest_id,string"`
	Legacy             APIV2Tweet       `json:"legacy"`
	Tombstone          *Tombstone       `json:"tombstone"`
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
	EditControl struct {
		EditTweetIDs Int64Slice `json:"edit_tweet_ids"`
	} `json:"edit_control"`
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

	if api_result.Result.Legacy.ID == 0 && api_result.Result.Tweet.Legacy.ID != 0 ||
		api_result.Result.ID == 0 && api_result.Result.Tweet.ID != 0 {
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
	main_tweet_trove, err := api_result.Result.Legacy.ToTweetTrove()
	if errors.Is(err, ERR_NO_TWEET) {
		// If the tweet is edited, the entry is just a list of the more recent versions
		edit_tweet_ids := api_result.Result.EditControl.EditTweetIDs
		if api_result.Result.ID != 0 && len(edit_tweet_ids) > 1 && edit_tweet_ids[len(edit_tweet_ids)-1] != api_result.Result.ID {
			// There's a more recent version of the tweet available
			main_tweet_trove.Tweets[TweetID(api_result.Result.ID)] = Tweet{
				TombstoneType: "newer-version-available",
				ID:            TweetID(api_result.Result.ID),
			}
		} else {
			// Not edited; something else is wrong
			return TweetTrove{}, err
		}
	} else if err != nil {
		panic(err)
	}
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

		// Handle `"quoted_status_result": {}` results
		if errors.Is(err, ERR_NO_TWEET) {
			// Replace it with a tombstone
			err = ErrorIsTombstone
			if quoted_api_result.Result.Tombstone == nil {
				quoted_api_result.Result.Tombstone = &Tombstone{}
			}
			quoted_api_result.Result.Tombstone.Text.Text = "This Post is unavailable. Learn more"
		}

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
		main_tweet, is_ok := ret.Tweets[TweetID(api_result.Result.ID)]
		if !is_ok {
			panic(fmt.Errorf("Tweet trove didn't contain its own tweet with ID %d:\n  %w", api_result.Result.ID, EXTERNAL_API_ERROR))
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

func (api_v2_tweet APIV2Tweet) ToTweetTrove() (TweetTrove, error) {
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
			return ret, fmt.Errorf("parsing APIV2Tweet: %w", err)
		}
		ret.Tweets[main_tweet.ID] = main_tweet
	}

	return ret, nil
}

type ItemContent struct {
	ItemType     string      `json:"itemType"`
	TweetResults APIV2Result `json:"tweet_results"`
	APIV2UserResult

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
			log.Warnf("Panic while decoding entry: %s\n", e.OriginalJSON)
			panic(obj)
		}
	}()
	if e.Content.EntryType == "TimelineTimelineCursor" || e.Content.ItemContent.ItemType == "TimelineTimelineCursor" {
		// Ignore cursor entries.
		// - e.Content.EntryType -> User Feed itself
		// - e.Content.ItemContent.ItemType -> conversation thread in a user feed
		return NewTweetTrove()
	} else if e.Content.ItemContent.ItemType == "TimelineLabel" {
		// Skip inline "labels" like "More Replies" that appear when you click "show more replies"
		return NewTweetTrove()
	} else if e.Content.EntryType == "TimelineTimelineModule" {
		ret := NewTweetTrove()

		parts := strings.Split(e.EntryID, "-")
		if parts[0] == "homeConversation" || parts[0] == "conversationthread" ||
			strings.Join(parts[0:2], "-") == "profile-conversation" || strings.Join(parts[0:2], "-") == "home-conversation" {
			// Process it.
			// - "profile-conversation": conversation thread on a user feed
			// - "homeConversation": This looks like it got changed to "profile-conversation"
			// - "home-conversation": probably same as above lol-- someone did some refactoring
			// - "conversationthread": conversation thread in the replies under a TweetDetail view
			for _, item := range e.Content.Items {
				if item.Item.ItemContent.ItemType == "TimelineTimelineCursor" {
					// "Show More" replies button in a thread on Tweet Detail page
					continue
				}
				if item.Item.ItemContent.ItemType == "TimelineTweetComposer" {
					// Composer button
					continue
				}
				trove, err := item.Item.ItemContent.TweetResults.ToTweetTrove()
				if errors.Is(err, ErrorIsTombstone) {
					// TODO: do something with tombstones in replies to a Tweet Detail
					// For now, just ignore tombstones in the replies
				} else if err != nil {
					panic(err)
				}
				ret.MergeWith(trove)
			}
		} else if parts[0] == "whoToFollow" || parts[0] == "TopicsModule" || parts[0] == "tweetdetailrelatedtweets" {
			// Ignore "Who to follow", "Topics" and "Related Tweets" modules.
			// TODO: maybe we can capture these eventually
			log.Debugf("Skipping %s entry", e.EntryID)
		} else {
			log.Warn("TimelineTimelineModule with unknown EntryID: " + e.EntryID)
		}

		return ret
	} else if e.Content.EntryType == "TimelineTimelineItem" {
		if e.Content.ItemContent.ItemType == "TimelineTombstone" {
			// TODO: user feed tombstone entries
			return NewTweetTrove()
		}
		if strings.Split(e.EntryID, "-")[0] == "messageprompt" {
			return NewTweetTrove()
		}
		ret, err := e.Content.ItemContent.TweetResults.ToTweetTrove()

		// Handle tombstones in the focused tweet and parent reply thread
		if errors.Is(err, ErrorIsTombstone) {
			ret = NewTweetTrove() // clear the result just in case there is a TweetID(0) in it
			tombstoned_tweet := APITweet{}

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

			fake_user := GetUnknownUser()
			ret.Users[fake_user.ID] = fake_user
			parsed_tombstone_tweet.UserID = fake_user.ID
			ret.Tweets[parsed_tombstone_tweet.ID] = parsed_tombstone_tweet
		} else if err != nil {
			if e.Content.ItemContent.APIV2UserResult.UserResults.Result.ID != 0 {
				user := e.Content.ItemContent.APIV2UserResult.ToUser()
				ret = NewTweetTrove()
				ret.Users[user.ID] = user
			} else {
				panic(err)
			}
		}
		return ret
	}
	panic("Unknown EntryType: " + e.Content.EntryType)
}

type APIV2Instruction struct {
	Type    string       `json:"type"`
	Entries []APIV2Entry `json:"entries"`
	Entry   APIV2Entry   `json:"entry"`
}

type APIV2Response struct {
	Data struct {
		Home struct {
			HomeTimelineUrt struct {
				Instructions []APIV2Instruction `json:"instructions"`
			} `json:"home_timeline_urt"`
		} `json:"home"`
		User struct {
			Result struct {
				TimelineV2 struct { // "Likes" feed calls this "timeline_v2" for some reason
					Timeline struct {
						Instructions []APIV2Instruction `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline_v2"`
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
		SearchByRawQuery struct {
			SearchTimeline struct {
				Timeline struct {
					Instructions []APIV2Instruction `json:"instructions"`
				} `json:"timeline"`
			} `json:"search_timeline"`
		} `json:"search_by_raw_query"`
		BookmarkTimelineV2 struct {
			Timeline struct {
				Instructions []APIV2Instruction `json:"instructions"`
			} `json:"timeline"`
		} `json:"bookmark_timeline_v2"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

func (api_response APIV2Response) GetMainInstruction() *APIV2Instruction {
	instructions := api_response.Data.User.Result.Timeline.Timeline.Instructions
	for i := range instructions {
		if instructions[i].Type == "TimelineAddEntries" {
			return &instructions[i]
		}
	}
	instructions = api_response.Data.User.Result.TimelineV2.Timeline.Instructions
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
	instructions = api_response.Data.Home.HomeTimelineUrt.Instructions
	for i := range instructions {
		if instructions[i].Type == "TimelineAddEntries" {
			return &instructions[i]
		}
	}
	instructions = api_response.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions
	for i := range instructions {
		if instructions[i].Type == "TimelineAddEntries" {
			return &instructions[i]
		}
	}
	instructions = api_response.Data.BookmarkTimelineV2.Timeline.Instructions
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
		if entry.Content.ItemContent.CursorType == "Bottom" ||
			entry.Content.ItemContent.CursorType == "ShowMoreThreadsPrompt" ||
			entry.Content.ItemContent.CursorType == "ShowMoreThreads" {
			// "Bottom": normal cursor, auto-loads when it scrolls into view
			// "ShowMoreThreads": normal cursor, but you have to click it to load more
			// "ShowMoreThreadsPrompt": show offensive/low quality replies
			return entry.Content.ItemContent.Value
		}
	}
	return ""
}

// Returns `true` if there's no non-cursor entries in this response, false otherwise
func (api_response APIV2Response) IsEmpty() bool {
	for _, e := range api_response.GetMainInstruction().Entries {
		if !strings.Contains(e.EntryID, "cursor") {
			return false
		}
	}
	return true
}

// Parse the collected API response and turn it into a TweetTrove
func (api_response APIV2Response) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()

	// Parse all of the entries, and attempt to do tombstone reply-joining as we go
	for i, entry := range api_response.GetMainInstruction().Entries { // TODO: the second Instruction is the pinned tweet in a User Feed
		ret.MergeWith(entry.ToTweetTrove())

		// Only do tombstone reply-joining on a Tweet Detail thread (not a User Feed!)
		if len(api_response.Data.ThreadedConversationWithInjectionsV2.Instructions) == 0 {
			continue
		}
		// Skip the first entry since it doesn't have a parent
		if i == 0 {
			continue
		}
		// Infer "in_reply_to_id" for tombstoned tweets from the order of entries, if applicable
		if entry.Content.EntryType == "TimelineTimelineItem" {
			entry_type, main_tweet_id := entry.ParseID()
			if entry_type == "cursor-showmorethreadsprompt" ||
				entry_type == "cursor-bottom" ||
				entry_type == "cursor-showmorethreads" ||
				entry_type == "cursor-top" {
				// Skip cursors
				// - "cursor-top" => So far, the only top-cursor type there is
				// - "cursor-bottom" => auto-loads more replies when you scroll it into view
				// - "cursor-showmorethreadsprompt" => "Show additional replies, including those that may contain offensive content" button
				// - "cursor-showmorethreads" => "Show more replies" button
				continue
			}
			if entry_type == "label" {
				// Skip labels / headers
				continue
			}
			if entry_type != "tweet" {
				// TODO: discovery panic
				panic(fmt.Sprintf("Unexpected first part of entry id: %q", entry_type))
			}
			main_tweet, is_ok := ret.Tweets[main_tweet_id]
			if !is_ok {
				// On a User Feed, the entry ID could also be a Retweet ID, but we should only reply-join on Tweet Detail views.
				panic(fmt.Sprintf("Entry didn't parse correctly: %q", entry.EntryID))
			}
			if main_tweet.InReplyToID != TweetID(0) {
				// Already has an InReplyToID, so doesn't need to be joined using positional inference
				continue
			}
			_, prev_entry_id := api_response.GetMainInstruction().Entries[i-1].ParseID()
			main_tweet.InReplyToID = prev_entry_id
			ret.Tweets[main_tweet_id] = main_tweet
		}
		// else if entry.Content.EntryType == "TimelineTimelineModule" {
		// 	// TODO: check reply threads for tombstones as well
		// }
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

		// Fill out the replied tweet's UserID using this tweet's "in_reply_to_user_id".
		// If this tweet doesn't have it (i.e., this tweet is also a tombstone), create a fake user instead, and add it to the tweet trove.
		if replied_tweet.UserID == 0 || replied_tweet.UserID == GetUnknownUser().ID {
			replied_tweet.UserID = tweet.in_reply_to_user_id
			if replied_tweet.UserID == 0 || replied_tweet.UserID == GetUnknownUser().ID {
				fake_user := GetUnknownUser()
				ret.Users[fake_user.ID] = fake_user
				replied_tweet.UserID = fake_user.ID
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
		ret.TombstoneUsers = append(ret.TombstoneUsers, existing_user.Handle)

		ret.Tweets[replied_tweet.ID] = replied_tweet
	}

	return ret, nil // TODO: This doesn't need to return an error, it's always nil
}

func (r APIV2Response) GetPinnedTweetAsTweetTrove() TweetTrove {
	for _, instr := range r.Data.User.Result.Timeline.Timeline.Instructions {
		if instr.Type == "TimelinePinEntry" {
			return instr.Entry.ToTweetTrove()
		}
	}
	for _, instr := range r.Data.User.Result.TimelineV2.Timeline.Instructions {
		if instr.Type == "TimelinePinEntry" {
			return instr.Entry.ToTweetTrove()
		}
	}
	return NewTweetTrove()
}

func (r APIV2Response) ToTweetTroveAsLikes() (TweetTrove, error) {
	ret, err := r.ToTweetTrove()
	if err != nil {
		return ret, err
	}

	// Post-process tweets as Likes
	for _, entry := range r.GetMainInstruction().Entries {
		// Skip cursors
		if entry.Content.EntryType == "TimelineTimelineCursor" {
			continue
		}
		// Assume it's not a TimelineModule or a Tombstone
		if entry.Content.EntryType != "TimelineTimelineItem" {
			panic(fmt.Sprintf("Unknown Like entry type: %s", entry.Content.EntryType))
		}
		if entry.Content.ItemContent.ItemType == "TimelineTombstone" {
			panic(fmt.Sprintf("Liked tweet is a tombstone: %#v", entry))
		}

		// Generate a "Like" from the entry
		tweet, is_ok := ret.Tweets[TweetID(entry.Content.ItemContent.TweetResults.Result._Result.ID)]
		if !is_ok {
			// For TweetWithVisibilityResults
			tweet, is_ok = ret.Tweets[TweetID(entry.Content.ItemContent.TweetResults.Result.Tweet.ID)]
			if !is_ok {
				log.Warnf("ID: %d", entry.Content.ItemContent.TweetResults.Result._Result.ID)
				log.Warnf("Entry JSON: %s", entry.OriginalJSON)
				panic(ret.Tweets)
			}
		}
		ret.Likes[LikeSortID(entry.SortIndex)] = Like{
			SortID:  LikeSortID(entry.SortIndex),
			TweetID: tweet.ID,
		}
	}
	return ret, err
}

func (r APIV2Response) ToTweetTroveAsBookmarks() (TweetTrove, error) {
	ret, err := r.ToTweetTrove()
	if err != nil {
		return ret, err
	}

	// Post-process tweets as Bookmarks
	for _, entry := range r.GetMainInstruction().Entries {
		// Skip cursors
		if entry.Content.EntryType == "TimelineTimelineCursor" {
			continue
		}
		// Assume it's not a TimelineModule or a Tombstone
		if entry.Content.EntryType != "TimelineTimelineItem" {
			panic(fmt.Sprintf("Unknown Bookmark entry type: %s", entry.Content.EntryType))
		}
		if entry.Content.ItemContent.ItemType == "TimelineTombstone" {
			panic(fmt.Sprintf("Bookmarkd tweet is a tombstone: %#v", entry))
		}

		// Generate a "Bookmark" from the entry
		tweet, is_ok := ret.Tweets[TweetID(entry.Content.ItemContent.TweetResults.Result._Result.ID)]
		if !is_ok {
			// For TweetWithVisibilityResults
			tweet, is_ok = ret.Tweets[TweetID(entry.Content.ItemContent.TweetResults.Result.Tweet.ID)]
			if !is_ok {
				log.Warnf("ID: %d", entry.Content.ItemContent.TweetResults.Result._Result.ID)
				log.Warnf("Entry JSON: %s", entry.OriginalJSON)
				panic(ret.Tweets)
			}
		}
		ret.Bookmarks[BookmarkSortID(entry.SortIndex)] = Bookmark{
			SortID:  BookmarkSortID(entry.SortIndex),
			TweetID: tweet.ID,
		}
	}
	return ret, err
}

type PaginatedQuery interface {
	NextPage(api *API, cursor string) (APIV2Response, error)
	ToTweetTrove(r APIV2Response) (TweetTrove, error)
}

func (api *API) GetMore(pq PaginatedQuery, response *APIV2Response, count int) error {
	last_response := response
	for last_response.GetCursorBottom() != "" && len(response.GetMainInstruction().Entries) < count {
		fresh_response, err := pq.NextPage(api, last_response.GetCursorBottom())
		if err != nil {
			return fmt.Errorf("error getting next page for %#v: %w", pq, err)
		}

		if fresh_response.GetCursorBottom() == last_response.GetCursorBottom() && len(fresh_response.GetMainInstruction().Entries) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			fmt.Printf("Cursor repeated; EOF\n")
			return END_OF_FEED
		}
		if fresh_response.IsEmpty() {
			// Response has a pinned tweet, but no other content: end of feed has been reached
			fmt.Printf("No non-pinned-tweet entries; EOF\n")
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

func (api *API) GetPaginatedQuery(pq PaginatedQuery, count int) (TweetTrove, error) {
	fmt.Printf("Paginating %d count\n", count)
	api_response, err := pq.NextPage(api, "")
	if err != nil {
		return TweetTrove{}, fmt.Errorf("Error calling API to fetch query %#v:\n  %w", pq, err)
	}
	if len(api_response.GetMainInstruction().Entries) < count && api_response.GetCursorBottom() != "" {
		err = api.GetMore(pq, &api_response, count)
		if errors.Is(err, END_OF_FEED) {
			println("End of feed!")
		} else if err != nil {
			return TweetTrove{}, err
		}
	}

	trove, err := pq.ToTweetTrove(api_response)
	if err != nil {
		return TweetTrove{}, fmt.Errorf("Error parsing the tweet trove for query %#v:\n  %w", pq, err)
	}

	fmt.Println("------------")
	err = trove.PostProcess()
	return trove, err
}

// Get a User feed using the new GraphQL twitter api
func (api *API) GetGraphqlFeedFor(user_id UserID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/Q6aAvPw7azXZbqXzuqTALA/UserTweetsAndReplies",
		Variables: GraphqlVariables{
			UserID:                      user_id,
			Count:                       40,
			Cursor:                      cursor,
			IncludePromotedContent:      false,
			WithCommunity:               true,
			WithSuperFollowsUserFields:  true,
			WithBirdwatchPivots:         false,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithBirdwatchNotes:          false,
			WithVoice:                   true,
			WithV2Timeline:              true,
		},
		Features: GraphqlFeatures{
			ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
			CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			LongformNotetweetsConsumptionEnabled:                           true,
			FreedomOfSpeechNotReachFetchEnabled:                            false,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: true,
			LongformNotetweetsRichTextReadEnabled:                          true,
			LongformNotetweetsInlineMediaEnabled:                           false,
			ResponsiveWebMediaDownloadVideoEnabled:                         true,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response APIV2Response
	err = api.do_http(url.String(), cursor, &response)

	return response, err
}

type PaginatedUserFeed struct {
	user_id UserID
}

func (p PaginatedUserFeed) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetGraphqlFeedFor(p.user_id, cursor)
}
func (p PaginatedUserFeed) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	ret, err := r.ToTweetTrove()
	// Add the pinned tweet, if applicable
	ret.MergeWith(r.GetPinnedTweetAsTweetTrove())
	return ret, err
}

func (api *API) GetTweetDetail(tweet_id TweetID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/tPRAv4UnqM9dOgDWggph7Q/TweetDetail",
		Variables: GraphqlVariables{
			FocalTweetID:                           tweet_id,
			Cursor:                                 cursor,
			WithRuxInjections:                      false,
			IncludePromotedContent:                 false,
			WithCommunity:                          true,
			WithQuickPromoteEligibilityTweetFields: true,
			WithBirdwatchNotes:                     true,
			WithVoice:                              true,
			WithV2Timeline:                         true,
		},
		Features: GraphqlFeatures{
			RWebListsTimelineRedesignEnabled:                               true,
			ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
			VerifiedPhoneLabelEnabled:                                      false,
			CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      false,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			ViewCountsEverywhereApiEnabled:                                 true,
			LongformNotetweetsConsumptionEnabled:                           true,
			TweetAwardsWebTippingEnabled:                                   false,
			FreedomOfSpeechNotReachFetchEnabled:                            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			LongformNotetweetsRichTextReadEnabled:                          true,
			LongformNotetweetsInlineMediaEnabled:                           false,
			ResponsiveWebEnhanceCardsEnabled:                               false,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response APIV2Response
	err = api.do_http(url.String(), cursor, &response)
	if len(response.Errors) != 0 {
		if response.Errors[0].Message == "_Missing: No status found with that ID." {
			return response, ErrDoesntExist
		}
		return response, fmt.Errorf("%w: %s", EXTERNAL_API_ERROR, response.Errors[0].Message)
	}

	return response, err
}

type PaginatedTweetReplies struct {
	tweet_id TweetID
}

func (p PaginatedTweetReplies) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetTweetDetail(p.tweet_id, cursor)
}
func (p PaginatedTweetReplies) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	return r.ToTweetTrove()
}

func (api *API) GetUserLikes(user_id UserID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/2Z6LYO4UTM4BnWjaNCod6g/Likes",
		Variables: GraphqlVariables{
			UserID:                      user_id,
			Count:                       20,
			Cursor:                      cursor,
			IncludePromotedContent:      false,
			WithSuperFollowsUserFields:  true,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithBirdwatchNotes:          false,
			WithVoice:                   true,
			WithV2Timeline:              false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			VerifiedPhoneLabelEnabled:                                      false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled:  true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebTextConversationsEnabled:                          false,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response APIV2Response
	err = api.do_http(url.String(), cursor, &response)
	if err != nil {
		panic(err)
	}
	return response, nil
}

type PaginatedUserLikes struct {
	user_id UserID
}

func (p PaginatedUserLikes) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetUserLikes(p.user_id, cursor)
}
func (p PaginatedUserLikes) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	ret, err := r.ToTweetTroveAsLikes()
	if err != nil {
		return TweetTrove{}, err
	}

	// Fill out the liking UserID
	for i := range ret.Likes {
		l := ret.Likes[i]
		l.UserID = p.user_id
		ret.Likes[i] = l
	}
	return ret, nil
}

func GetUserLikes(user_id UserID, how_many int) (TweetTrove, error) {
	return the_api.GetPaginatedQuery(PaginatedUserLikes{user_id}, how_many)
}

func (api *API) GetBookmarks(cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/xLjCVTqYWz8CGSprLU349w/Bookmarks",
		Variables: GraphqlVariables{
			Count:                  20,
			Cursor:                 cursor,
			IncludePromotedContent: false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled:  true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebEnhanceCardsEnabled:                               true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			ResponsiveWebTextConversationsEnabled:                          false,
			VerifiedPhoneLabelEnabled:                                      false,

			CommunitiesWebEnableTweetCommunityResultsFetch: true,
			RWebTipjarConsumptionEnabled:                   true,
			ArticlesPreviewEnabled:                         true,
			GraphqlTimelineV2BookmarkTimeline:              true,
			CreatorSubscriptionsQuoteTweetPreviewEnabled:   false,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response APIV2Response
	err = api.do_http(url.String(), cursor, &response)
	if err != nil {
		panic(err)
	}
	return response, nil
}

type PaginatedBookmarks struct {
	user_id UserID
}

func (p PaginatedBookmarks) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetBookmarks(cursor)
}
func (p PaginatedBookmarks) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	ret, err := r.ToTweetTroveAsBookmarks()
	if err != nil {
		return TweetTrove{}, err
	}

	// Fill out the bookmarking UserID
	for i := range ret.Bookmarks {
		l := ret.Bookmarks[i]
		l.UserID = p.user_id
		ret.Bookmarks[i] = l
	}
	return ret, nil
}

func GetBookmarks(how_many int) (TweetTrove, error) {
	return the_api.GetPaginatedQuery(PaginatedBookmarks{the_api.UserID}, how_many)
}

func (api *API) GetHomeTimeline(cursor string, is_following_only bool) (TweetTrove, error) {
	var url string
	body_struct := struct {
		Variables GraphqlVariables `json:"variables"`
		Features  GraphqlFeatures  `json:"features"`
		QueryID   string           `json:"queryId"`
	}{
		Variables: GraphqlVariables{
			Count:                  40,
			Cursor:                 cursor,
			IncludePromotedContent: false,
			// LatestControlAvailable: true, // TODO: new field?
			WithCommunity: true,
			// SeenTweetIDs: []string{"...some TweetIDs"}? // TODO: new field?
		},
		Features: GraphqlFeatures{
			RWebListsTimelineRedesignEnabled:                               true,
			ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
			VerifiedPhoneLabelEnabled:                                      false,
			CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      false,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			ViewCountsEverywhereApiEnabled:                                 true,
			LongformNotetweetsConsumptionEnabled:                           true,
			TweetAwardsWebTippingEnabled:                                   false,
			FreedomOfSpeechNotReachFetchEnabled:                            false,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: true,
			LongformNotetweetsRichTextReadEnabled:                          true,
			LongformNotetweetsInlineMediaEnabled:                           true,
			ResponsiveWebEnhanceCardsEnabled:                               false,
			ResponsiveWebTwitterArticleTweetConsumptionEnabled:             false,
			ResponsiveWebMediaDownloadVideoEnabled:                         false,
		},
	}
	if is_following_only {
		body_struct.QueryID = "iMKdg5Vq-ldwmiqCbvX1QA"
		url = "https://twitter.com/i/api/graphql/iMKdg5Vq-ldwmiqCbvX1QA/HomeLatestTimeline"
	} else {
		body_struct.QueryID = "W4Tpu1uueTGK53paUgxF0Q"
		url = "https://twitter.com/i/api/graphql/W4Tpu1uueTGK53paUgxF0Q/HomeTimeline"
	}
	var response APIV2Response
	body_bytes, err := json.Marshal(body_struct)
	if err != nil {
		panic(err)
	}
	err = api.do_http_POST(url, string(body_bytes), &response)
	if err != nil {
		panic(err)
	}
	trove, err := response.ToTweetTrove()
	if err != nil {
		return TweetTrove{}, err
	}
	return trove, err
}

func GetHomeTimeline(cursor string, is_following_only bool) (TweetTrove, error) {
	return the_api.GetHomeTimeline(cursor, is_following_only)
}

func (api API) GetUser(handle UserHandle) (APIUser, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://api.twitter.com/graphql/SAMkL5y_N9pmahSw8yy6gw/UserByScreenName",
		Variables: GraphqlVariables{
			ScreenName:                  handle,
			Count:                       20,
			IncludePromotedContent:      false,
			WithSuperFollowsUserFields:  true,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithBirdwatchNotes:          false,
			WithVoice:                   true,
			WithV2Timeline:              false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			VerifiedPhoneLabelEnabled:                                      false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled:  true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebTextConversationsEnabled:                          false,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response UserResponse
	err = api.do_http(url.String(), "", &response)
	if err != nil {
		panic(err)
	}

	return response.ConvertToAPIUser(), nil
}

func (api *API) Search(query string, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/NA567V_8AFwu0cZEkAAKcw/SearchTimeline",
		Variables: GraphqlVariables{
			RawQuery:                    query,
			Count:                       50,
			Product:                     "Top",
			Cursor:                      cursor,
			IncludePromotedContent:      false,
			WithSuperFollowsUserFields:  true,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithBirdwatchNotes:          false,
			WithVoice:                   true,
			WithV2Timeline:              false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			VerifiedPhoneLabelEnabled:                                      false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled:  true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebTextConversationsEnabled:                          false,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var result APIV2Response
	err = api.do_http(url.String(), cursor, &result)
	return result, err
}

type PaginatedSearch struct {
	query string
}

func (p PaginatedSearch) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.Search(p.query, cursor)
}
func (p PaginatedSearch) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	return r.ToTweetTrove()
}
