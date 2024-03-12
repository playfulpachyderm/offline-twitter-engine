package scraper

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type APIDMReaction struct {
	ID        int    `json:"id,string"`
	Time      int    `json:"time,string"`
	SenderID  int    `json:"sender_id,string"`
	Emoji     string `json:"emoji_reaction"`
	MessageID int    `json:"message_id,string"`
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
		Entities struct {
			URLs []struct {
				ExpandedURL  string `json:"expanded_url"`
				ShortenedUrl string `json:"url"`
			} `json:"urls"`
		} `json:"entities"`
		Attachment struct {
			Tweet struct {
				Url    string `json:"url"`
				Status struct {
					APITweet
					User APIUser `json:"user"`
				} `json:"status"`
			} `json:"tweet"`
			Photo APIMedia         `json:"photo"`
			Video APIExtendedMedia `json:"video"`
			Card  APICard          `json:"card"`
		} `json:"attachment"`
	} `json:"message_data"`
	MessageReactions []APIDMReaction `json:"message_reactions"`
}

// Remove embedded tweet short-URLs
func (m *APIDMMessage) NormalizeContent() {
	// All URLs
	for _, url := range m.MessageData.Entities.URLs {
		index := strings.Index(m.MessageData.Text, url.ShortenedUrl)
		if index == (len(m.MessageData.Text) - len(url.ShortenedUrl)) {
			m.MessageData.Text = strings.TrimSpace(m.MessageData.Text[0:index])
		}
	}

	// Specific items
	if m.MessageData.Attachment.Tweet.Status.ID != 0 {
		m.MessageData.Text = strings.Replace(m.MessageData.Text, m.MessageData.Attachment.Tweet.Url, "", 1)
	}
	if m.MessageData.Attachment.Photo.ID != 0 {
		m.MessageData.Text = strings.Replace(m.MessageData.Text, m.MessageData.Attachment.Photo.URL, "", 1)
	}
	if m.MessageData.Attachment.Video.ID != 0 {
		m.MessageData.Text = strings.Replace(m.MessageData.Text, m.MessageData.Attachment.Video.URL, "", 1)
	}

	m.MessageData.Text = strings.TrimSpace(m.MessageData.Text)
}

func (m APIDMMessage) ToDMTrove() DMTrove {
	ret := NewDMTrove()
	if m.ID == 0 {
		return ret
	}

	m.NormalizeContent()
	result := ParseAPIDMMessage(m)

	// Parse tweet attachment
	if m.MessageData.Attachment.Tweet.Status.ID != 0 {
		u, err := ParseSingleUser(m.MessageData.Attachment.Tweet.Status.User)
		if err != nil {
			panic(err)
		}
		ret.Users[u.ID] = u

		t, err := ParseSingleTweet(m.MessageData.Attachment.Tweet.Status.APITweet)
		if err != nil {
			panic(err)
		}
		t.UserID = u.ID
		ret.Tweets[t.ID] = t
		result.EmbeddedTweetID = t.ID
	}
	ret.Messages[result.ID] = result

	// TODO: parse attached images and videos

	return ret
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

	// For type == "GROUP_DM"
	CreateTime      int    `json:"create_time,string"`
	CreatedByUserID int    `json:"created_by_user_id,string"`
	Name            string `json:"name"`
	AvatarImage     string `json:"avatar_image_https"`
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
		Message          APIDMMessage  `json:"message"`
		ReactionCreate   APIDMReaction `json:"reaction_create"`
		JoinConversation struct {
			ID             int    `json:"id,string"`
			ConversationID string `json:"conversation_id"`
			SenderID       int    `json:"sender_id,string"`
			Time           int    `json:"time,string"`
			Participants   []struct {
				UserID int `json:"user_id,string"`
			} `json:"participants"`
		} `json:"join_conversation"`
		TrustConversation struct {
			ID             int    `json:"id,string"`
			ConversationID string `json:"conversation_id"`
			Reason         string `json:"reason"`
			Time           int    `json:"time,string"`
		} `json:"trust_conversation"`
		ParticipantsLeave struct {
			ID             int    `json:"id,string"`
			ConversationID string `json:"conversation_id"`
			Time           int    `json:"time,string"`
			Participants   []struct {
				UserID int `json:"user_id,string"`
			} `json:"participants"`
		} `json:"participants_leave"`
		ConversationRead struct {
			ID              int    `json:"id,string"`
			Time            int    `json:"time,string"`
			ConversationID  string `json:"conversation_id"`
			LastReadEventID int    `json:"last_read_event_id,string"`
		} `json:"conversation_read"`
	} `json:"entries"`
	Users         map[string]APIUser           `json:"users"`
	Conversations map[string]APIDMConversation `json:"conversations"`
}

type APIDMResponse struct {
	InboxInitialState    APIInbox `json:"inbox_initial_state"`
	InboxTimeline        APIInbox `json:"inbox_timeline"`
	ConversationTimeline APIInbox `json:"conversation_timeline"`
	UserEvents           APIInbox `json:"user_events"`
}

func (r APIInbox) ToDMTrove() DMTrove {
	ret := NewDMTrove()

	for _, entry := range r.Entries {
		if entry.JoinConversation.ID != 0 || entry.TrustConversation.ID != 0 ||
			entry.ParticipantsLeave.ID != 0 || entry.ConversationRead.ID != 0 {
			// TODO: message invitations
			// TODO: people join/leave the chat
			// TODO: updating read/unread indicators
			continue
		}
		if entry.ReactionCreate.ID != 0 {
			// Convert it into a Message
			entry.Message.ID = entry.ReactionCreate.MessageID
			entry.Message.MessageReactions = []APIDMReaction{entry.ReactionCreate}
		}

		// TODO:
		// if _, is_ok := ret.Messages[result.ID]; is_ok {
		// 	// No clobbering
		// 	panic("Already in the trove: " + fmt.Sprint(result.ID))
		// }

		ret.MergeWith(entry.Message.ToDMTrove())
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

func (api *API) GetDMConversation(id DMChatRoomID, max_id DMMessageID) (APIInbox, error) {
	url, err := url.Parse("https://twitter.com/i/api/1.1/dm/conversation/" + string(id) + ".json")
	if err != nil {
		panic(err)
	}
	query := url.Query()
	query.Add("max_id", fmt.Sprint(max_id))
	query.Add("context", "FETCH_DM_CONVERSATION_HISTORY")
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
	query.Add("include_ext_limited_action_results", "true")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "false")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("include_conversation_info", "true")
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
	return result.ConversationTimeline, err
}

func (api *API) PollInboxUpdates(cursor string) (APIInbox, error) {
	url, err := url.Parse("https://twitter.com/i/api/1.1/dm/user_updates.json")
	if err != nil {
		panic(err)
	}
	query := url.Query()
	query.Add("cursor", cursor)
	query.Add("nsfw_filtering_enabled", "false")
	query.Add("filter_low_quality", "true")
	query.Add("include_quality", "all")
	query.Add("dm_secret_conversations_enabled", "false")
	query.Add("krs_registration_enabled", "true")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_ext_limited_action_results", "true")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "false")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("include_ext_edit_control", "true")
	query.Add("include_ext_business_affiliations_label", "true")
	query.Add("ext", strings.Join([]string{
		"mediaColor",
		"altText",
		"businessAffiliationsLabel",
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
	return result.UserEvents, err
}

func (api *API) SendDMMessage(room_id DMChatRoomID, text string, in_reply_to_id DMMessageID) (APIInbox, error) {
	url, err := url.Parse("https://twitter.com/i/api/1.1/dm/new2.json")
	if err != nil {
		panic(err)
	}

	query := url.Query()
	query.Add("nsfw_filtering_enabled", "false")
	query.Add("filter_low_quality", "true")
	query.Add("include_quality", "all")
	query.Add("dm_secret_conversations_enabled", "false")
	query.Add("krs_registration_enabled", "true")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_ext_limited_action_results", "true")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "false")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("include_ext_edit_control", "true")
	query.Add("include_ext_business_affiliations_label", "true")
	query.Add("ext", strings.Join([]string{
		"mediaColor",
		"altText",
		"businessAffiliationsLabel",
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

	request_id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	replying_to_text := ""
	if in_reply_to_id != 0 {
		replying_to_text = fmt.Sprintf(`"reply_to_dm_id":"%d",`, in_reply_to_id)
	}

	post_data := `{"conversation_id":"` + string(room_id) +
		`","recipient_ids":false,"request_id":"` + request_id.String() +
		`","text":"` + text + `",` +
		replying_to_text + `"cards_platform":"Web-12","include_cards":1,"include_quote_count":true,"dm_users":false}`

	var result APIInbox
	err = api.do_http_POST(url.String(), post_data, &result)
	return result, err
}
