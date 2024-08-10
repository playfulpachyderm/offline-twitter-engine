package scraper

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
func GetInbox(how_many int) (TweetTrove, string, error) {
	return the_api.GetInbox(how_many)
}

func (api *API) GetConversation(id DMChatRoomID, max_id DMMessageID, how_many int) (TweetTrove, error) {
	if !api.IsAuthenticated {
		return TweetTrove{}, ErrLoginRequired
	}
	dm_response, err := api.GetDMConversation(id, max_id)
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToTweetTrove(api.UserID)
	oldest := trove.GetOldestMessage(id)
	for len(trove.Messages) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = api.GetDMConversation(id, oldest)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToTweetTrove(api.UserID)
		oldest = next_trove.GetOldestMessage(id)
		trove.MergeWith(next_trove)
	}

	return trove, nil
}
func GetConversation(id DMChatRoomID, max_id DMMessageID, how_many int) (TweetTrove, error) {
	return the_api.GetConversation(id, max_id, how_many)
}

func PollInboxUpdates(cursor string) (TweetTrove, string, error) {
	return the_api.PollInboxUpdates(cursor)
}

func SendDMMessage(room_id DMChatRoomID, text string, in_reply_to_id DMMessageID) (TweetTrove, error) {
	return the_api.SendDMMessage(room_id, text, in_reply_to_id)
}

func SendDMReaction(room_id DMChatRoomID, message_id DMMessageID, reacc string) error {
	return the_api.SendDMReaction(room_id, message_id, reacc)
}

func MarkDMChatRead(room_id DMChatRoomID, read_message_id DMMessageID) error {
	return the_api.MarkDMChatRead(room_id, read_message_id)
}
