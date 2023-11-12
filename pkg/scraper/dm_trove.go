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

func GetInbox() DMTrove {
	if !the_api.IsAuthenticated {
		log.Fatalf("Fetching DMs can only be done when authenticated.  Please provide `--session [user]`")
	}
	dm_response, err := the_api.GetDMInbox()
	if err != nil {
		panic(err)
	}
	return dm_response.ToDMTrove()
}
