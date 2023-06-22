package scraper

type DMTrove struct {
	Rooms      map[DMChatRoomID]DMChatRoom
	Messages   map[DMMessageID]DMMessage
	TweetTrove TweetTrove
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
