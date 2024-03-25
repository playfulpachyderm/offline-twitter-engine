package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

const (
	CHAT_MESSAGES_ALL_SQL_FIELDS = `id, chat_room_id, sender_id, sent_at, request_id, text, in_reply_to_id, embedded_tweet_id`
	CHAT_ROOMS_ALL_SQL_FIELDS    = `id, type, last_messaged_at, is_nsfw, created_at, created_by_user_id, name,
	                                             avatar_image_remote_url, avatar_image_local_path`
	CHAT_ROOM_PARTICIPANTS_ALL_SQL_FIELDS = `chat_room_id, user_id, last_read_event_id, is_chat_settings_valid, is_notifications_disabled,
	                                             is_mention_notifications_disabled, is_read_only, is_trusted, is_muted, status`
)

func (p Profile) SaveChatRoom(r DMChatRoom) error {
	_, err := p.DB.NamedExec(`
		insert into chat_rooms (id, type, last_messaged_at, is_nsfw, created_at, created_by_user_id, name,
		                        avatar_image_remote_url, avatar_image_local_path)
		                values (:id, :type, :last_messaged_at, :is_nsfw, :created_at, :created_by_user_id, :name,
		                        :avatar_image_remote_url, :avatar_image_local_path)
		 on conflict do update
		                   set last_messaged_at=:last_messaged_at,
		                       name=:name,
		                       avatar_image_remote_url=:avatar_image_remote_url,
		                       avatar_image_local_path=:avatar_image_local_path
		`, r,
	)
	if err != nil {
		return fmt.Errorf("Error executing SaveChatRoom(ID %s).  Info: %#v:\n  %w", r.ID, r, err)
	}

	for _, participant := range r.Participants {
		_, err = p.DB.NamedExec(`
		insert into chat_room_participants (
			chat_room_id,
			user_id,
			last_read_event_id,
			is_chat_settings_valid,
			is_notifications_disabled,
			is_mention_notifications_disabled,
			is_read_only,
			is_trusted,
			is_muted,
			status)
		values (
			:chat_room_id,
			:user_id,
			:last_read_event_id,
			:is_chat_settings_valid,
			:is_notifications_disabled,
			:is_mention_notifications_disabled,
			:is_read_only,
			:is_trusted,
			:is_muted,
			:status)
		 on conflict do update
		set last_read_event_id=:last_read_event_id,
			is_chat_settings_valid=:is_chat_settings_valid,
			is_notifications_disabled=:is_notifications_disabled,
			is_mention_notifications_disabled=:is_mention_notifications_disabled,
			is_read_only=:is_read_only,
			is_trusted=:is_trusted,
			is_muted=:is_muted,
			status=:status
		`, participant,
		)
	}
	if err != nil {
		return fmt.Errorf("Error saving chat participant: %#v\n  %w", r, err)
	}
	return nil
}

// Get a chat room with participants.
//
// Since this function is only used for tests and to confirm the room exists in `fetch_dm` and
// `send_dm` command-line subcommands, it doesn't need to be super efficient.  So just reuse the
// full DMChatRoom fetch function and throw away the stuff we don't need
func (p Profile) GetChatRoom(id DMChatRoomID) (room DMChatRoom, err error) {
	chat_view := p.GetChatRoomContents(id, 0)
	return chat_view.Rooms[id], nil
}

func (p Profile) SaveChatMessage(m DMMessage) error {
	// The message itself
	_, err := p.DB.NamedExec(`
		insert into chat_messages (id, chat_room_id, sender_id, sent_at, request_id, in_reply_to_id, text, embedded_tweet_id)
		values (:id, :chat_room_id, :sender_id, :sent_at, :request_id, :in_reply_to_id, :text, :embedded_tweet_id)
		on conflict do nothing
		`, m,
	)
	if err != nil {
		return fmt.Errorf("Error saving message: %#v\n  %w", m, err)
	}

	// Reactions
	for _, reacc := range m.Reactions {
		fmt.Println(reacc)
		_, err = p.DB.NamedExec(`
			insert into chat_message_reactions (id, message_id, sender_id, sent_at, emoji)
			values (:id, :message_id, :sender_id, :sent_at, :emoji)
			on conflict do nothing
			`, reacc,
		)
		if err != nil {
			return fmt.Errorf("Error saving message reaction (message %d, reacc %d): %#v\n  %w", m.ID, reacc.ID, reacc, err)
		}
	}

	// Images
	for _, img := range m.Images {
		_, err := p.DB.NamedExec(`
			insert into chat_message_images (id, chat_message_id, width, height, remote_url, local_filename, is_downloaded)
			            values (:id, :chat_message_id, :width, :height, :remote_url, :local_filename, :is_downloaded)
			       on conflict do update
			               set is_downloaded=(is_downloaded or :is_downloaded)
			`,
			img,
		)
		if err != nil {
			return fmt.Errorf("Error saving image (message ID %d):\n  %w", img.DMMessageID, err)
		}
	}

	// Videos
	for _, vid := range m.Videos {
		_, err := p.DB.NamedExec(`
			insert into chat_message_videos
			            (id, chat_message_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename,
			             duration, view_count, is_downloaded, is_blocked_by_dmca, is_gif)
			     values (:id, :chat_message_id, :width, :height, :remote_url, :local_filename, :thumbnail_remote_url,
			            :thumbnail_local_filename, :duration, :view_count, :is_downloaded, :is_blocked_by_dmca, :is_gif)
			on conflict do update
			        set is_downloaded=(is_downloaded or :is_downloaded),
			            view_count=max(view_count, :view_count),
			            is_blocked_by_dmca = :is_blocked_by_dmca
			`,
			vid,
		)
		if err != nil {
			return fmt.Errorf("Error saving video (message ID %d):\n  %w", vid.DMMessageID, err)
		}
	}

	// Urls
	for _, url := range m.Urls {
		_, err := p.DB.NamedExec(`
			insert into chat_message_urls (chat_message_id, domain, text, short_text, title, description, creator_id, site_id,
			                               thumbnail_width, thumbnail_height, thumbnail_remote_url, thumbnail_local_path, has_card,
			                               has_thumbnail, is_content_downloaded)
			     values (:chat_message_id, :domain, :text, :short_text, :title, :description, :creator_id, :site_id, :thumbnail_width,
			             :thumbnail_height, :thumbnail_remote_url, :thumbnail_local_path, :has_card, :has_thumbnail, :is_content_downloaded
			            )
			on conflict do update
			        set is_content_downloaded=(is_content_downloaded or :is_content_downloaded)
		`, url)
		if err != nil {
			return fmt.Errorf("Error saving Url (message ID %d):\n  %w", url.DMMessageID, err)
		}
	}

	return nil
}

// Get a single chat message, filling its attachment contents.
//
// This function is only used in tests.
func (p Profile) GetChatMessage(id DMMessageID) (ret DMMessage, err error) {
	err = p.DB.Get(&ret, `
		select `+CHAT_MESSAGES_ALL_SQL_FIELDS+`
		  from chat_messages
		 where id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat message %d:\n  %w", id, err)
	}
	ret.Reactions = make(map[UserID]DMReaction)

	// This is a bit circuitous, but it doesn't matter because this function is only used in tests
	trove := NewDMTrove()
	trove.Messages[ret.ID] = ret
	p.fill_dm_contents(&trove)

	return trove.Messages[ret.ID], nil
}

type DMChatView struct {
	DMTrove
	RoomIDs      []DMChatRoomID
	MessageIDs   []DMMessageID
	ActiveRoomID DMChatRoomID
}

func NewDMChatView() DMChatView {
	return DMChatView{
		DMTrove:    NewDMTrove(),
		RoomIDs:    []DMChatRoomID{},
		MessageIDs: []DMMessageID{},
	}
}

// Get the list of chat rooms the given user is in, including participants and latest message preview
func (p Profile) GetChatRoomsPreview(id UserID) DMChatView {
	ret := NewDMChatView()

	// Get the list of rooms
	var rooms []DMChatRoom
	err := p.DB.Select(&rooms, `
		select `+CHAT_ROOMS_ALL_SQL_FIELDS+`
		  from chat_rooms
		 where exists (select 1 from chat_room_participants where chat_room_id = chat_rooms.id and user_id = ?)
		 order by last_messaged_at desc
	`, id)
	if err != nil {
		panic(err)
	}

	// Fill data for the rooms
	for _, room := range rooms {
		// Fetch the latest message
		var msg DMMessage
		q, args, err := sqlx.Named(`
			select `+CHAT_MESSAGES_ALL_SQL_FIELDS+`
			  from chat_messages
			 where chat_room_id = :room_id
			   and sent_at = (select max(sent_at) from chat_messages where chat_room_id = :room_id)
		`, struct {
			ID DMChatRoomID `db:"room_id"`
		}{ID: room.ID})
		if err != nil {
			panic(err)
		}
		err = p.DB.Get(&msg, q, args...)
		if errors.Is(err, sql.ErrNoRows) {
			// TODO
			fmt.Printf("No messages found in chat; skipping preview\n")
		} else if err != nil {
			panic(err)
		}

		// Fetch the participants
		p.fill_chat_room_participants(&room, &ret.DMTrove)

		// Add everything to the Trove
		room.LastMessageID = msg.ID
		ret.Rooms[room.ID] = room
		ret.Messages[msg.ID] = msg
		ret.RoomIDs = append(ret.RoomIDs, room.ID)
	}
	return ret
}

// Get chat room detail, including participants and messages
func (p Profile) GetChatRoomContents(id DMChatRoomID, latest_timestamp int) DMChatView {
	ret := NewDMChatView()
	var room DMChatRoom
	err := p.DB.Get(&room, `
		select `+CHAT_ROOMS_ALL_SQL_FIELDS+`
		  from chat_rooms
		 where id = ?
	`, id)
	if err != nil {
		panic(err)
	}

	// Fetch all messages
	var msgs []DMMessage
	err = p.DB.Select(&msgs, `
		select `+CHAT_MESSAGES_ALL_SQL_FIELDS+`
		  from chat_messages
		 where chat_room_id = ?
		   and sent_at > ?
		 order by sent_at desc
		 limit 50
	`, room.ID, latest_timestamp)
	if err != nil {
		panic(err)
	}
	ret.MessageIDs = make([]DMMessageID, len(msgs))
	for i, msg := range msgs {
		ret.MessageIDs[len(ret.MessageIDs)-i-1] = msg.ID // Reverse order
		msg.Reactions = make(map[UserID]DMReaction)
		ret.Messages[msg.ID] = msg
	}

	// Fetch the participants
	p.fill_chat_room_participants(&room, &ret.DMTrove)

	// Set last message ID on chat room
	if len(ret.MessageIDs) > 0 {
		// If there's no messages, it should be OK to have LastMessageID = 0, since this is only used
		// to generate previews
		room.LastMessageID = ret.MessageIDs[len(ret.MessageIDs)-1]
	}

	// Put the room in the Trove
	ret.Rooms[room.ID] = room

	p.fill_dm_contents(&ret.DMTrove)
	return ret
}

// Fetch the chat participants and insert it into the DMChatRoom.  Inserts user information
// into the DMTrove.
func (p Profile) fill_chat_room_participants(room *DMChatRoom, trove *DMTrove) {
	var participants []struct {
		DMChatParticipant
		User
	}
	err := p.DB.Select(&participants, `
		select `+CHAT_ROOM_PARTICIPANTS_ALL_SQL_FIELDS+`, `+USERS_ALL_SQL_FIELDS+`
		  from chat_room_participants join users on chat_room_participants.user_id = users.id
		 where chat_room_id = ?
	`, room.ID)
	if err != nil {
		panic(err)
	}
	room.Participants = make(map[UserID]DMChatParticipant)
	for _, p := range participants {
		room.Participants[p.User.ID] = p.DMChatParticipant
		trove.Users[p.User.ID] = p.User
	}
}

// Fetch reaccs, attachments/embeds and replied-to messages and add them to the DMTrove
func (p Profile) fill_dm_contents(trove *DMTrove) {
	// Skip processing if there's no messages whomst'd've contents to fetch
	if len(trove.Messages) == 0 {
		return
	}

	// Fetch all reaccs
	var reaccs []DMReaction
	message_ids := []interface{}{}
	for _, msg := range trove.Messages {
		message_ids = append(message_ids, msg.ID)
	}
	err := p.DB.Select(&reaccs, `
		select id, message_id, sender_id, sent_at, emoji
		  from chat_message_reactions
		 where message_id in (`+strings.Repeat("?,", len(trove.Messages)-1)+`?)
	`, message_ids...)
	if err != nil {
		panic(err)
	}
	for _, reacc := range reaccs {
		msg := trove.Messages[reacc.DMMessageID]
		msg.Reactions[reacc.SenderID] = reacc
		trove.Messages[reacc.DMMessageID] = msg
	}

	// Images
	var images []Image
	err = p.DB.Select(&images, `
		select id, chat_message_id, width, height, remote_url, local_filename, is_downloaded
		  from chat_message_images
		 where chat_message_id in (`+strings.Repeat("?,", len(trove.Messages)-1)+`?)
	`, message_ids...)
	if err != nil {
		panic(err)
	}
	for _, img := range images {
		msg := trove.Messages[img.DMMessageID]
		msg.Images = []Image{img}
		trove.Messages[msg.ID] = msg
	}

	// Videos
	var videos []Video
	err = p.DB.Select(&videos, `
		select id, chat_message_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename,
		       duration, view_count, is_downloaded, is_blocked_by_dmca, is_gif
		  from chat_message_videos
		 where chat_message_id in (`+strings.Repeat("?,", len(trove.Messages)-1)+`?)
	`, message_ids...)
	if err != nil {
		panic(err)
	}
	for _, vid := range videos {
		println("asdfasfasdf")
		msg := trove.Messages[vid.DMMessageID]
		msg.Videos = []Video{vid}
		trove.Messages[msg.ID] = msg
	}

	// Urls
	var urls []Url
	err = p.DB.Select(&urls, `
		select chat_message_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
		       thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded
		  from chat_message_urls
		 where chat_message_id in (`+strings.Repeat("?,", len(trove.Messages)-1)+`?)
	`, message_ids...)
	if err != nil {
		panic(err)
	}
	for _, url := range urls {
		msg := trove.Messages[url.DMMessageID]
		msg.Urls = []Url{url}
		trove.Messages[msg.ID] = msg
	}

	// Fetch all embedded tweets
	embedded_tweet_ids := []interface{}{}
	for _, m := range trove.Messages {
		if m.EmbeddedTweetID != 0 {
			embedded_tweet_ids = append(embedded_tweet_ids, m.EmbeddedTweetID)
		}
	}
	if len(embedded_tweet_ids) > 0 {
		var embedded_tweets []Tweet
		err = p.DB.Select(&embedded_tweets, `
		     select `+TWEETS_ALL_SQL_FIELDS+`
		       from tweets
		  left join tombstone_types on tweets.tombstone_type = tombstone_types.rowid
		  left join likes on tweets.id = likes.tweet_id and likes.user_id = ?
		      where id in (`+strings.Repeat("?,", len(embedded_tweet_ids)-1)+`?)`,
			append([]interface{}{UserID(0)}, embedded_tweet_ids...)...)
		if err != nil {
			panic(err)
		}
		for _, t := range embedded_tweets {
			trove.Tweets[t.ID] = t
		}
	}

	// Fetch replied-to message previews
	replied_message_ids := []interface{}{}
	for _, m := range trove.Messages {
		if m.InReplyToID != 0 {
			// Don't clobber if it's already been fetched
			if _, is_ok := trove.Messages[m.InReplyToID]; !is_ok {
				replied_message_ids = append(replied_message_ids, m.InReplyToID)
			}
		}
	}
	if len(replied_message_ids) > 0 {
		var replied_msgs []DMMessage
		err = p.DB.Select(&replied_msgs, `
			select `+CHAT_MESSAGES_ALL_SQL_FIELDS+`
			  from chat_messages
		     where id in (`+strings.Repeat("?,", len(replied_message_ids)-1)+`?)`,
			replied_message_ids...)
		if err != nil {
			panic(err)
		}
		for _, msg := range replied_msgs {
			msg.Reactions = make(map[UserID]DMReaction)
			trove.Messages[msg.ID] = msg
		}
	}

	p.fill_content(&trove.TweetTrove, UserID(0))
}


	}
}
