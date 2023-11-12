package persistence

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (p Profile) SaveChatRoom(r DMChatRoom) error {
	_, err := p.DB.NamedExec(`
		insert into chat_rooms (id, type, last_messaged_at, is_nsfw)
					    values (:id, :type, :last_messaged_at, :is_nsfw)
         on conflict do update
                           set last_messaged_at=:last_messaged_at
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
	// }
	return nil
}

func (p Profile) GetChatRoom(id DMChatRoomID) (ret DMChatRoom, err error) {
	err = p.DB.Get(&ret, `
        select id, type, last_messaged_at, is_nsfw
          from chat_rooms
         where id = ?
    `, id)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat room (%s):\n  %w", id, err)
	}

	participants := []DMChatParticipant{}
	err = p.DB.Select(&participants, `
		select chat_room_id, user_id, last_read_event_id, is_chat_settings_valid, is_notifications_disabled,
		       is_mention_notifications_disabled, is_read_only, is_trusted, is_muted, status
		  from chat_room_participants
		 where chat_room_id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat room participants (%s):\n  %w", id, err)
	}
	ret.Participants = make(map[UserID]DMChatParticipant)
	for _, p := range participants {
		ret.Participants[p.UserID] = p
	}
	return ret, nil
}

func (p Profile) SaveChatMessage(m DMMessage) error {
	_, err := p.DB.NamedExec(`
		insert into chat_messages (id, chat_room_id, sender_id, sent_at, request_id, in_reply_to_id, text)
		values (:id, :chat_room_id, :sender_id, :sent_at, :request_id, :in_reply_to_id, :text)
		on conflict do nothing
		`, m,
	)
	if err != nil {
		return fmt.Errorf("Error saving message: %#v\n  %w", m, err)
	}

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
	return nil
}

func (p Profile) GetChatMessage(id DMMessageID) (ret DMMessage, err error) {
	err = p.DB.Get(&ret, `
		select id, chat_room_id, sender_id, sent_at, request_id, text, in_reply_to_id
		  from chat_messages
		 where id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat message (%d):\n  %w", id, err)
	}

	reaccs := []DMReaction{}
	err = p.DB.Select(&reaccs, `
		select id, message_id, sender_id, sent_at, emoji
		  from chat_message_reactions
		 where message_id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting reactions to chat message (%d):\n  %w", id, err)
	}
	ret.Reactions = make(map[UserID]DMReaction)
	for _, r := range reaccs {
		ret.Reactions[r.SenderID] = r
	}
	return ret, nil
}

type DMChatView struct {
	DMTrove
	RoomIDs    []DMChatRoomID
	MessageIDs []DMMessageID
}

func NewDMChatView() DMChatView {
	return DMChatView{
		DMTrove:    NewDMTrove(),
		RoomIDs:    []DMChatRoomID{},
		MessageIDs: []DMMessageID{},
	}
}

func (p Profile) GetChatRoomsPreview(id UserID) DMChatView {
	ret := NewDMChatView()

	var rooms []DMChatRoom
	err := p.DB.Select(&rooms, `
		select id, type, last_messaged_at, is_nsfw
		  from chat_rooms
		 where exists (select 1 from chat_room_participants where chat_room_id = chat_rooms.id and user_id = ?)
		 order by last_messaged_at desc
	`, id)
	if err != nil {
		panic(err)
	}
	for _, room := range rooms {
		// Fetch the latest message
		var msg DMMessage
		q, args, err := sqlx.Named(`
			select id, chat_room_id, sender_id, sent_at, request_id, text, in_reply_to_id
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
		if err != nil {
			panic(err)
		}

		// Fetch the participants
		// DUPE chat-room-participants-SQL
		var participants []struct {
			DMChatParticipant
			User
		}
		err = p.DB.Select(&participants, `
			select chat_room_id, user_id, last_read_event_id, is_chat_settings_valid, is_notifications_disabled,
			       is_mention_notifications_disabled, is_read_only, is_trusted, is_muted, status, `+USERS_ALL_SQL_FIELDS+`
		      from chat_room_participants join users on chat_room_participants.user_id = users.id
		     where chat_room_id = ?
		`, room.ID)
		if err != nil {
			panic(err)
		}
		room.Participants = make(map[UserID]DMChatParticipant)
		for _, participant := range participants {
			room.Participants[participant.User.ID] = participant.DMChatParticipant
			ret.Users[participant.User.ID] = participant.User
		}

		// Add everything to the Trove
		room.LastMessageID = msg.ID
		ret.Rooms[room.ID] = room
		ret.Messages[msg.ID] = msg
		ret.RoomIDs = append(ret.RoomIDs, room.ID)
	}
	return ret
}

func (p Profile) GetChatRoomContents(id DMChatRoomID) DMChatView {
	ret := NewDMChatView()
	var room DMChatRoom
	err := p.DB.Get(&room, `
		select id, type, last_messaged_at, is_nsfw
	      from chat_rooms
	     where id = ?
	`, id)
	if err != nil {
		panic(err)
	}

	// Fetch the participants
	// DUPE chat-room-participants-SQL
	var participants []struct {
		DMChatParticipant
		User
	}
	err = p.DB.Select(&participants, `
		select chat_room_id, user_id, last_read_event_id, is_chat_settings_valid, is_notifications_disabled,
		       is_mention_notifications_disabled, is_read_only, is_trusted, is_muted, status, `+USERS_ALL_SQL_FIELDS+`
	      from chat_room_participants join users on chat_room_participants.user_id = users.id
	     where chat_room_id = ?
	`, room.ID)
	if err != nil {
		panic(err)
	}
	room.Participants = make(map[UserID]DMChatParticipant)
	for _, participant := range participants {
		room.Participants[participant.User.ID] = participant.DMChatParticipant
		ret.Users[participant.User.ID] = participant.User
	}

	// Fetch all messages
	var msgs []DMMessage
	err = p.DB.Select(&msgs, `
		select id, chat_room_id, sender_id, sent_at, request_id, text, in_reply_to_id
	      from chat_messages
	     where chat_room_id = :room_id
	     order by sent_at desc
	     limit 50
	`, room.ID)
	if err != nil {
		panic(err)
	}
	ret.MessageIDs = make([]DMMessageID, len(msgs))
	for i, msg := range msgs {
		ret.MessageIDs[len(ret.MessageIDs)-i-1] = msg.ID
		msg.Reactions = make(map[UserID]DMReaction)
		ret.Messages[msg.ID] = msg
	}

	// Set last message ID on chat room
	room.LastMessageID = ret.MessageIDs[len(ret.MessageIDs)-1]

	// Put the room in the Trove
	ret.Rooms[room.ID] = room

	// Fetch all reaccs
	var reaccs []DMReaction
	message_ids_copy := make([]interface{}, len(ret.MessageIDs))
	for i, id := range ret.MessageIDs {
		message_ids_copy[i] = id
	}
	err = p.DB.Select(&reaccs, `
	    select id, message_id, sender_id, sent_at, emoji
	      from chat_message_reactions
		 where message_id in (`+strings.Repeat("?,", len(ret.MessageIDs)-1)+`?)
	`, message_ids_copy...)
	if err != nil {
		panic(err)
	}
	for _, reacc := range reaccs {
		msg := ret.Messages[reacc.DMMessageID]
		msg.Reactions[reacc.SenderID] = reacc
		ret.Messages[reacc.DMMessageID] = msg
	}

	return ret
}
