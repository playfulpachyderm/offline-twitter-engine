package webserver

import (
	"net/http"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type MessageData persistence.DMChatView

func (t MessageData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t MessageData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}
func (t MessageData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}
func (t MessageData) Space(id scraper.SpaceID) scraper.Space {
	return t.Spaces[id]
}
func (t MessageData) FocusedTweetID() scraper.TweetID {
	return scraper.TweetID(0)
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Messages' handler (path: %q)", r.URL.Path)

	// TODO: what if no active user?

	chat_view := app.Profile.GetChatRoomsPreview(app.ActiveUser.ID)
	if strings.Trim(r.URL.Path, "/") != "" {
		message_id := scraper.DMChatRoomID(strings.Trim(r.URL.Path, "/"))
		chat_contents := app.Profile.GetChatRoomContents(message_id)
		chat_view.MergeWith(chat_contents.DMTrove)
		chat_view.MessageIDs = chat_contents.MessageIDs

		if r.Header.Get("HX-Request") == "true" {
			app.buffered_render_tweet_htmx(w, "chat-view", MessageData(chat_view))
			return
		}
	}

	app.buffered_render_tweet_page(w, "tpl/messages.tpl", MessageData(chat_view))
}

// type DMChatView struct {
// 	scraper.DMTrove
// 	RoomIDs    []scraper.DMChatRoomID
// 	MessageIDs []scraper.DMMessageID
// }
