package persistence

import (
	"fmt"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Convenience function that saves all the objects in a TweetTrove.
// Panics if anything goes wrong.
func (p Profile) SaveDMTrove(trove DMTrove, should_download bool) {
	p.SaveTweetTrove(trove.TweetTrove, should_download)

	for _, r := range trove.Rooms {
		err := p.SaveChatRoom(r)
		if err != nil {
			panic(fmt.Errorf("Error saving chat room: %#v\n  %w", r, err))
		}
	}
	for _, m := range trove.Messages {
		err := p.SaveChatMessage(m)
		if err != nil {
			panic(fmt.Errorf("Error saving chat message: %#v\n  %w", m, err))
		}
	}
}
