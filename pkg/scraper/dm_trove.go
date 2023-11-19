package scraper

import (
	log "github.com/sirupsen/logrus"
)

type DMTrove struct {
	Rooms    map[DMChatRoomID]DMChatRoom
	Messages map[DMMessageID]DMMessage
	TweetTrove
}

func NewDMTrove() DMTrove {
	ret := DMTrove{}
	ret.Rooms = make(map[DMChatRoomID]DMChatRoom)
	ret.Messages = make(map[DMMessageID]DMMessage)
	ret.TweetTrove = NewTweetTrove()
	return ret
}

func (t1 *DMTrove) MergeWith(t2 DMTrove) {
	for id, val := range t2.Rooms {
		t1.Rooms[id] = val
	}
	for id, val := range t2.Messages {
		t1.Messages[id] = val
	}
	t1.TweetTrove.MergeWith(t2.TweetTrove)
}

// Returns a DMTrove and the cursor for the next update
func GetInbox(how_many int) (DMTrove, string) {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.GetDMInbox()
	if err != nil {
		panic(err)
	}

	trove := dm_response.ToDMTrove()
	cursor := dm_response.Cursor
	next_cursor_id := dm_response.InboxTimelines.Trusted.MinEntryID
	for len(trove.Rooms) < how_many && dm_response.Status != "AT_END" {
		dm_response, err = the_api.GetInboxTrusted(next_cursor_id)
		if err != nil {
			panic(err)
		}
		next_trove := dm_response.ToDMTrove()
		next_cursor_id = dm_response.MinEntryID
		trove.MergeWith(next_trove)
	}

	return trove, cursor
}
