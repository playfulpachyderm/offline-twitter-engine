package scraper

import (
	log "github.com/sirupsen/logrus"
)

func (t TweetTrove) GetOldestMessage(id DMChatRoomID) DMMessageID {
	oldest := DMMessageID(^uint(0) >> 1) // Max integer
	for _, m := range t.Messages {
		if m.ID < oldest && m.DMChatRoomID == id {
			oldest = m.ID
		}
	}
	return oldest
}

//  TODO: Why are these all here?  =>

// Returns a TweetTrove and the cursor for the next update
func GetInbox(how_many int) (TweetTrove, string) {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.GetDMInbox()
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToTweetTrove()
	cursor := dm_response.Cursor
	next_cursor_id := dm_response.InboxTimelines.Trusted.MinEntryID
	for len(trove.Rooms) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = the_api.GetInboxTrusted(next_cursor_id)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToTweetTrove()
		next_cursor_id = dm_response.MinEntryID
		trove.MergeWith(next_trove)
	}

	return trove, cursor
}

func GetConversation(id DMChatRoomID, max_id DMMessageID, how_many int) TweetTrove {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.GetDMConversation(id, max_id)
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToTweetTrove()
	oldest := trove.GetOldestMessage(id)
	for len(trove.Messages) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = the_api.GetDMConversation(id, oldest)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToTweetTrove()
		oldest = next_trove.GetOldestMessage(id)
		trove.MergeWith(next_trove)
	}

	return trove
}

// Returns a TweetTrove and the cursor for the next update
func PollInboxUpdates(cursor string) (TweetTrove, string) {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.PollInboxUpdates(cursor)
	if err != nil {
		panic(err)
	}

	return dm_response.ToTweetTrove(), dm_response.Cursor
}

func SendDMMessage(room_id DMChatRoomID, text string, in_reply_to_id DMMessageID) TweetTrove {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.SendDMMessage(room_id, text, in_reply_to_id)
	if err != nil {
		panic(err)
	}
	return dm_response.ToTweetTrove()
}
func SendDMReaction(room_id DMChatRoomID, message_id DMMessageID, reacc string) error {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	return the_api.SendDMReaction(room_id, message_id, reacc)
}
func MarkDMChatRead(room_id DMChatRoomID, read_message_id DMMessageID) {
	if !the_api.IsAuthenticated {
		log.Fatalf("Writing DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	the_api.MarkDMChatRead(room_id, read_message_id)
}
