package persistence

import (
	"fmt"

	"offline_twitter/scraper"
)

func (p Profile) SaveChatRoom(r scraper.DMChatRoom) error {
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

func (p Profile) GetChatRoom(id scraper.DMChatRoomID) (ret scraper.DMChatRoom, err error) {
	err = p.DB.Get(&ret, `
        select id, type, last_messaged_at, is_nsfw
          from chat_rooms
         where id = ?
    `, id)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat room (%s):\n  %w", id, err)
	}

	participants := []scraper.DMChatParticipant{}
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
	ret.Participants = make(map[scraper.UserID]scraper.DMChatParticipant)
	for _, p := range participants {
		ret.Participants[p.UserID] = p
	}
	return ret, nil
}

func (p Profile) SaveChatMessage(m scraper.DMMessage) error {
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

func (p Profile) GetChatMessage(id scraper.DMMessageID) (ret scraper.DMMessage, err error) {
	err = p.DB.Get(&ret, `
		select id, chat_room_id, sender_id, sent_at, request_id, text, in_reply_to_id
		  from chat_messages
		 where id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting chat message (%d):\n  %w", id, err)
	}

	reaccs := []scraper.DMReaction{}
	err = p.DB.Select(&reaccs, `
		select id, message_id, sender_id, sent_at, emoji
		  from chat_message_reactions
		 where message_id = ?
		`, id,
	)
	if err != nil {
		return ret, fmt.Errorf("Error getting reactions to chat message (%d):\n  %w", id, err)
	}
	ret.Reactions = make(map[scraper.UserID]scraper.DMReaction)
	for _, r := range reaccs {
		ret.Reactions[r.SenderID] = r
	}
	return ret, nil
}
