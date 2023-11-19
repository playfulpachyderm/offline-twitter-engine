package scraper

import (
	"fmt"
	"net/url"
	"strings"
)

type APIDMReaction struct {
	ID       int    `json:"id,string"`
	Time     int    `json:"time,string"`
	SenderID int    `json:"sender_id,string"`
	Emoji    string `json:"emoji_reaction"`
}

type APIDMMessage struct {
	ID             int    `json:"id,string"`
	Time           int    `json:"time,string"`
	ConversationID string `json:"conversation_id"`
	MessageData    struct {
		ID        int    `json:"id,string"`
		Time      int    `json:"time,string"`
		SenderID  int    `json:"sender_id,string"`
		Text      string `json:"text"`
		ReplyData struct {
			ID int `json:"id,string"`
		} `json:"reply_data"`
		Attachment struct {
			Tweet APITweet `json:"tweet"`
		} `json:"attachment"`
	} `json:"message_data"`
	MessageReactions []APIDMReaction `json:"message_reactions"`
}

type APIDMConversation struct {
	ConversationID string `json:"conversation_id"`
	Type           string `json:"type"`
	SortTimestamp  int    `json:"sort_timestamp,string"`
	Participants   []struct {
		UserID          int `json:"user_id,string"`
		LastReadEventID int `json:"last_read_event_id,string"`
	}
	NSFW                  bool   `json:"nsfw"`
	NotificationsDisabled bool   `json:"notifications_disabled"`
	ReadOnly              bool   `json:"read_only"`
	Trusted               bool   `json:"trusted"`
	Muted                 bool   `json:"muted"`
	Status                string `json:"status"`
}

type APIInbox struct {
	Status          string `json:"status"`
	MinEntryID      int    `json:"min_entry_id,string"`
	LastSeenEventID int    `json:"last_seen_event_id,string"`
	Cursor          string `json:"cursor"`
	InboxTimelines  struct {
		Trusted struct {
			Status     string `json:"status"`
			MinEntryID int    `json:"min_entry_id,string"`
		} `json:"trusted"`
	} `json:"inbox_timelines"`
	Entries []struct {
		Message APIDMMessage `json:"message"`
	} `json:"entries"`
	Users         map[string]APIUser           `json:"users"`
	Conversations map[string]APIDMConversation `json:"conversations"`
}

type APIDMResponse struct {
	InboxInitialState APIInbox `json:"inbox_initial_state"`
	InboxTimeline     APIInbox `json:"inbox_timeline"`
}

func (r APIInbox) ToDMTrove() DMTrove {
	ret := NewDMTrove()

	for _, entry := range r.Entries {
		result := ParseAPIDMMessage(entry.Message)
		ret.Messages[result.ID] = result
		// TODO: parse Tweet attachments
	}
	for _, room := range r.Conversations {
		result := ParseAPIDMChatRoom(room)
		ret.Rooms[result.ID] = result
	}
	for _, u := range r.Users {
		result, err := ParseSingleUser(u)
		if err != nil {
			panic(err)
		}
		ret.TweetTrove.Users[result.ID] = result
	}
	return ret
}

func (api *API) GetDMInbox() (APIInbox, error) {
	url, err := url.Parse("https://twitter.com/i/api/1.1/dm/inbox_initial_state.json")
	if err != nil {
		panic(err)
	}
	query := url.Query()
	query.Add("nsfw_filtering_enabled", "false")
	query.Add("filter_low_quality", "true")
	query.Add("include_quality", "all")
	query.Add("include_profile_interstitial_type", "1")
	query.Add("include_blocking", "1")
	query.Add("include_blocked_by", "1")
	query.Add("include_followed_by", "1")
	query.Add("include_want_retweets", "1")
	query.Add("include_mute_edge", "1")
	query.Add("include_can_dm", "1")
	query.Add("include_can_media_tag", "1")
	query.Add("include_ext_has_nft_avatar", "1")
	query.Add("include_ext_is_blue_verified", "1")
	query.Add("include_ext_verified_type", "1")
	query.Add("include_ext_profile_image_shape", "1")
	query.Add("skip_status", "1")
	query.Add("dm_secret_conversations_enabled", "false")
	query.Add("krs_registration_enabled", "true")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_ext_limited_action_results", "false")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "true")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("include_ext_edit_control", "true")
	query.Add("ext", strings.Join([]string{
		"mediaColor",
		"altText",
		"mediaStats",
		"highlightedLabel",
		"hasNftAvatar",
		"voiceInfo",
		"birdwatchPivot",
		"enrichments",
		"superFollowMetadata",
		"unmentionInfo",
		"editControl",
		"vibe",
	}, ","))
	url.RawQuery = query.Encode()

	var result APIDMResponse
	err = api.do_http(url.String(), "", &result)
	result.InboxInitialState.Status = result.InboxInitialState.InboxTimelines.Trusted.Status
	return result.InboxInitialState, err
}

func (api *API) GetInboxTrusted(oldest_id int) (APIInbox, error) {
	url, err := url.Parse("https://twitter.com/i/api/1.1/dm/inbox_timeline/trusted.json")
	if err != nil {
		panic(err)
	}
	query := url.Query()
	query.Add("max_id", fmt.Sprint(oldest_id))
	query.Add("nsfw_filtering_enabled", "false")
	query.Add("filter_low_quality", "true")
	query.Add("include_quality", "all")
	query.Add("include_profile_interstitial_type", "1")
	query.Add("include_blocking", "1")
	query.Add("include_blocked_by", "1")
	query.Add("include_followed_by", "1")
	query.Add("include_want_retweets", "1")
	query.Add("include_mute_edge", "1")
	query.Add("include_can_dm", "1")
	query.Add("include_can_media_tag", "1")
	query.Add("include_ext_has_nft_avatar", "1")
	query.Add("include_ext_is_blue_verified", "1")
	query.Add("include_ext_verified_type", "1")
	query.Add("include_ext_profile_image_shape", "1")
	query.Add("skip_status", "1")
	query.Add("dm_secret_conversations_enabled", "false")
	query.Add("krs_registration_enabled", "true")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_ext_limited_action_results", "false")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "true")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("include_ext_edit_control", "true")
	query.Add("ext", strings.Join([]string{
		"mediaColor",
		"altText",
		"mediaStats",
		"highlightedLabel",
		"hasNftAvatar",
		"voiceInfo",
		"birdwatchPivot",
		"enrichments",
		"superFollowMetadata",
		"unmentionInfo",
		"editControl",
		"vibe",
	}, ","))
	url.RawQuery = query.Encode()

	var result APIDMResponse
	err = api.do_http(url.String(), "", &result)
	return result.InboxTimeline, err
}
