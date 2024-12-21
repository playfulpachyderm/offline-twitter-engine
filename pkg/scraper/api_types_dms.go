package scraper

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"path"
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

	m.MessageData.Text = html.UnescapeString(m.MessageData.Text)
	m.MessageData.Text = strings.TrimSpace(m.MessageData.Text)
}

func (api_msg APIDMMessage) ToTweetTrove() TweetTrove {
	ret := NewTweetTrove()
	if api_msg.ID == 0 {
		return ret
	}

	api_msg.NormalizeContent()

	msg := DMMessage{}
	msg.ID = DMMessageID(api_msg.ID)
	msg.SentAt = TimestampFromUnixMilli(int64(api_msg.Time))
	msg.DMChatRoomID = DMChatRoomID(api_msg.ConversationID)
	msg.SenderID = UserID(api_msg.MessageData.SenderID)
	msg.Text = api_msg.MessageData.Text

	msg.InReplyToID = DMMessageID(api_msg.MessageData.ReplyData.ID) // Will be "0" if not a reply

	msg.Reactions = make(map[UserID]DMReaction)
	for _, api_reacc := range api_msg.MessageReactions {
		reacc := DMReaction{}
		reacc.ID = DMMessageID(api_reacc.ID)
		reacc.SenderID = UserID(api_reacc.SenderID)
		reacc.SentAt = TimestampFromUnixMilli(int64(api_reacc.Time))
		reacc.Emoji = api_reacc.Emoji
		reacc.DMMessageID = msg.ID
		msg.Reactions[reacc.SenderID] = reacc
	}
	if api_msg.MessageData.Attachment.Photo.ID != 0 {
		new_image := ParseAPIMedia(api_msg.MessageData.Attachment.Photo)
		new_image.DMMessageID = msg.ID
		msg.Images = []Image{new_image}
	}
	if api_msg.MessageData.Attachment.Video.ID != 0 {
		entity := api_msg.MessageData.Attachment.Video
		if entity.Type == "video" || entity.Type == "animated_gif" {
			new_video := ParseAPIVideo(entity)
			new_video.DMMessageID = msg.ID
			msg.Videos = append(msg.Videos, new_video)
		}
	}

	// Process URLs and link previews
	for _, url := range api_msg.MessageData.Entities.URLs {
		// Skip it if it's an embedded tweet
		_, id, is_ok := TryParseTweetUrl(url.ExpandedURL)
		if is_ok && id == TweetID(api_msg.MessageData.Attachment.Tweet.Status.ID) {
			continue
		}
		// Skip it if it's an embedded image
		if api_msg.MessageData.Attachment.Photo.URL == url.ShortenedUrl {
			continue
		}
		// Skip it if it's an embedded video
		if api_msg.MessageData.Attachment.Video.URL == url.ShortenedUrl {
			continue
		}

		var new_url Url
		if api_msg.MessageData.Attachment.Card.ShortenedUrl == url.ShortenedUrl {
			if api_msg.MessageData.Attachment.Card.Name == "3691233323:audiospace" {
				// This "url" is just a link to a Space.  Don't process it as a Url
				// TODO: ...but do process it as a Space?
				continue
			}
			new_url = ParseAPIUrlCard(api_msg.MessageData.Attachment.Card)
		}
		new_url.Text = url.ExpandedURL
		new_url.ShortText = url.ShortenedUrl
		new_url.DMMessageID = msg.ID
		msg.Urls = append(msg.Urls, new_url)
	}

	// Parse tweet attachment
	if api_msg.MessageData.Attachment.Tweet.Status.ID != 0 {
		u, err := ParseSingleUser(api_msg.MessageData.Attachment.Tweet.Status.User)
		if err != nil {
			panic(err)
		}
		ret.Users[u.ID] = u

		t, err := ParseSingleTweet(api_msg.MessageData.Attachment.Tweet.Status.APITweet)
		if err != nil {
			panic(err)
		}
		t.UserID = u.ID
		ret.Tweets[t.ID] = t
		msg.EmbeddedTweetID = t.ID
	}
	ret.Messages[msg.ID] = msg

	return ret
}

type APIDMResponse struct {
	InboxInitialState    APIInbox `json:"inbox_initial_state"`
	InboxTimeline        APIInbox `json:"inbox_timeline"`
	ConversationTimeline APIInbox `json:"conversation_timeline"`
	UserEvents           APIInbox `json:"user_events"`
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

func (r APIInbox) ToTweetTrove(current_user_id UserID) TweetTrove {
	ret := NewTweetTrove()

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

		ret.MergeWith(entry.Message.ToTweetTrove())
	}
	for _, api_room := range r.Conversations {
		result := ParseAPIDMChatRoom(api_room, current_user_id)
		ret.Rooms[result.ID] = result
	}
	for _, u := range r.Users {
		result, err := ParseSingleUser(u)
		if err != nil {
			panic(err)
		}
		ret.Users[result.ID] = result
	}
	return ret
}

func ParseAPIDMChatRoom(api_room APIDMConversation, current_user_id UserID) DMChatRoom {
	result := DMChatRoom{}
	result.ID = DMChatRoomID(api_room.ConversationID)
	result.Type = api_room.Type
	result.LastMessagedAt = TimestampFromUnixMilli(int64(api_room.SortTimestamp))
	result.IsNSFW = api_room.NSFW

	if result.Type == "GROUP_DM" {
		result.CreatedAt = TimestampFromUnixMilli(int64(api_room.CreateTime))
		result.CreatedByUserID = UserID(api_room.CreatedByUserID)
		result.Name = api_room.Name
		result.AvatarImageRemoteURL = api_room.AvatarImage
		tmp_url, err := url.Parse(result.AvatarImageRemoteURL)
		if err != nil {
			panic(err)
		}
		result.AvatarImageLocalPath = fmt.Sprintf("%s_avatar_%s.%s", result.ID, path.Base(tmp_url.Path), tmp_url.Query().Get("format"))
	}

	result.Participants = make(map[UserID]DMChatParticipant)
	for _, api_participant := range api_room.Participants {
		participant := DMChatParticipant{}
		participant.UserID = UserID(api_participant.UserID)
		participant.DMChatRoomID = result.ID
		participant.LastReadEventID = DMMessageID(api_participant.LastReadEventID)

		// Process chat settings if this is the logged-in user
		if participant.UserID == current_user_id {
			participant.IsNotificationsDisabled = api_room.NotificationsDisabled
			participant.IsReadOnly = api_room.ReadOnly
			participant.IsTrusted = api_room.Trusted
			participant.IsMuted = api_room.Muted
			participant.Status = api_room.Status
			participant.IsChatSettingsValid = true
		}
		result.Participants[participant.UserID] = participant
	}
	return result
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
func (api *API) GetInbox(how_many int) (TweetTrove, string, error) {
	if !api.IsAuthenticated {
		return TweetTrove{}, "", ErrLoginRequired
	}
	dm_response, err := api.GetDMInbox()
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToTweetTrove(api.UserID)
	cursor := dm_response.Cursor
	next_cursor_id := dm_response.InboxTimelines.Trusted.MinEntryID
	for len(trove.Rooms) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = api.GetInboxTrusted(next_cursor_id)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToTweetTrove(api.UserID)
		next_cursor_id = dm_response.MinEntryID
		trove.MergeWith(next_trove)
	}

	return trove, cursor, nil
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

func (api *API) GetConversation(room_id DMChatRoomID, max_id DMMessageID, how_many int) (TweetTrove, error) {
	if !api.IsAuthenticated {
		return TweetTrove{}, ErrLoginRequired
	}

	fetch := func(max_id DMMessageID) (APIInbox, error) {
		url, err := url.Parse("https://twitter.com/i/api/1.1/dm/conversation/" + string(room_id) + ".json")
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

	dm_response, err := fetch(max_id)
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToTweetTrove(api.UserID)
	oldest := trove.GetOldestMessage(room_id)
	for len(trove.Messages) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = fetch(oldest)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToTweetTrove(api.UserID)
		oldest = next_trove.GetOldestMessage(room_id)
		trove.MergeWith(next_trove)
	}

	return trove, nil
}

// Returns a TweetTrove and the cursor for the next update, or an error
func (api *API) PollInboxUpdates(cursor string) (TweetTrove, string, error) {
	if !api.IsAuthenticated {
		return TweetTrove{}, "", ErrLoginRequired
	}
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
	if err != nil {
		return TweetTrove{}, "", err
	}
	return result.UserEvents.ToTweetTrove(api.UserID), result.UserEvents.Cursor, nil
}

// Writes
// ------

func (api *API) SendDMMessage(room_id DMChatRoomID, text string, in_reply_to_id DMMessageID) (TweetTrove, error) {
	if !api.IsAuthenticated {
		return TweetTrove{}, ErrLoginRequired
	}
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

	// Format safely as JSON (escape quotes, etc)
	sanitized_text, err := json.Marshal(text)
	if err != nil {
		panic(err)
	}

	post_data := `{"conversation_id":"` + string(room_id) +
		`","recipient_ids":false,"request_id":"` + request_id.String() +
		`","text":` + string(sanitized_text) + `,` +
		replying_to_text + `"cards_platform":"Web-12","include_cards":1,"include_quote_count":true,"dm_users":false}`

	var result APIInbox
	err = api.do_http_POST(url.String(), post_data, &result)

	if err != nil {
		return TweetTrove{}, err
	}
	return result.ToTweetTrove(api.UserID), nil
}

// Send a reacc
func (api *API) SendDMReaction(room_id DMChatRoomID, message_id DMMessageID, reacc string) error {
	if !api.IsAuthenticated {
		return ErrLoginRequired
	}
	url := "https://twitter.com/i/api/graphql/VyDyV9pC2oZEj6g52hgnhA/useDMReactionMutationAddMutation"
	body := `{"variables":{"conversationId":"` + string(room_id) + `","messageId":"` + fmt.Sprint(message_id) +
		`","reactionTypes":["Emoji"],"emojiReactions":["` + reacc + `"]},"queryId":"VyDyV9pC2oZEj6g52hgnhA"}`
	type SendDMResponse struct {
		Data struct {
			CreateDmReaction struct {
				Typename string `json:"__typename"`
			} `json:"create_dm_reaction"`
		} `json:"data"`
	}
	var result SendDMResponse
	err := api.do_http_POST(url, body, &result)
	if err != nil {
		return fmt.Errorf("Error executing HTTP POST:\n  %w", err)
	}
	if result.Data.CreateDmReaction.Typename != "CreateDMReactionSuccess" {
		return fmt.Errorf("Unexpected result sending DM reaction: %s", result.Data.CreateDmReaction.Typename)
	}
	return nil
}

// Mark a chat as read.
func (api *API) MarkDMChatRead(room_id DMChatRoomID, read_message_id DMMessageID) error {
	if !api.IsAuthenticated {
		return ErrLoginRequired
	}
	url := fmt.Sprintf("https://twitter.com/i/api/1.1/dm/conversation/%s/mark_read.json", room_id)

	// `do_http_POST` will set the "content-type" header based on whether the body starts with '{' or not.
	data := fmt.Sprintf("conversationId=%s&last_read_event_id=%d", room_id, read_message_id)

	return api.do_http_POST(url, data, nil) // Expected: HTTP 204
}
