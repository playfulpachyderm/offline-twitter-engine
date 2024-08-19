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
