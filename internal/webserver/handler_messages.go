package webserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"strconv"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type MessageData struct {
	persistence.DMChatView
	LatestPollingTimestamp int
}

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

	if r.URL.Query().Has("scrape") {
		// TODO: where is this going to be used?
		app.background_dm_polling_scrape()
	}

	chat_view_data := MessageData{DMChatView: app.Profile.GetChatRoomsPreview(app.ActiveUser.ID)} // Get message list previews

	if room_id != "" {
		// First send a message, if applicable
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
		chat_view_data.ActiveRoomID = room_id
		chat_view_data.LatestPollingTimestamp = -1
		if latest_timestamp_str := r.URL.Query().Get("latest_timestamp"); latest_timestamp_str != "" {
			var err error
			chat_view_data.LatestPollingTimestamp, err = strconv.Atoi(latest_timestamp_str)
			panic_if(err)
		}
		chat_contents := app.Profile.GetChatRoomContents(room_id, chat_view_data.LatestPollingTimestamp)
		chat_view_data.MergeWith(chat_contents.DMTrove)
		chat_view_data.MessageIDs = chat_contents.MessageIDs
		if len(chat_view_data.MessageIDs) > 0 {
			last_message_id := chat_view_data.MessageIDs[len(chat_view_data.MessageIDs) - 1]
			chat_view_data.LatestPollingTimestamp = int(chat_view_data.Messages[last_message_id].SentAt.Unix())
		}

		if r.URL.Query().Has("poll") {
			app.buffered_render_tweet_htmx(w, "messages-with-poller", chat_view_data)
			return
		}
	}

	app.buffered_render_tweet_page(w, "tpl/messages.tpl", chat_view_data)
}
