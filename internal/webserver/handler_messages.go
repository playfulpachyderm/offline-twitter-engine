package webserver

import (
	"encoding/json"
	"io"
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

	if app.ActiveUser.ID == 0 {
		app.error_401(w)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	room_id := scraper.DMChatRoomID(parts[0])
	if len(parts) == 2 && parts[1] == "send" {
		body, err := io.ReadAll(r.Body)
		panic_if(err)
		var message_data struct {
			Text string `json:"text"`
		}
		panic_if(json.Unmarshal(body, &message_data))
		trove := scraper.SendDMMessage(room_id, message_data.Text, 0)
		app.Profile.SaveDMTrove(trove, false)
		go app.Profile.SaveDMTrove(trove, true)
	}

	chat_view := app.Profile.GetChatRoomsPreview(app.ActiveUser.ID)
	if strings.Trim(r.URL.Path, "/") != "" {
		chat_view.ActiveRoomID = room_id
		chat_contents := app.Profile.GetChatRoomContents(room_id)
		chat_view.MergeWith(chat_contents.DMTrove)
		chat_view.MessageIDs = chat_contents.MessageIDs
	}

	app.buffered_render_tweet_page(w, "tpl/messages.tpl", MessageData(chat_view))
}
