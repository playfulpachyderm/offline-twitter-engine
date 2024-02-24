package webserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type MessageData struct {
	persistence.DMChatView
	LatestPollingTimestamp int
	ScrollBottom           bool
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Messages' handler (path: %q)", r.URL.Path)

	if app.ActiveUser.ID == 0 {
		app.error_401(w)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	room_id := scraper.DMChatRoomID(parts[0])

	if r.URL.Query().Has("poll") {
		// Not run as a goroutine; this call blocks.  It's not actually "background"
		app.background_dm_polling_scrape()
	}

	chat_view_data := MessageData{DMChatView: app.Profile.GetChatRoomsPreview(app.ActiveUser.ID)} // Get message list previews
	global_data := PageGlobalData{TweetTrove: chat_view_data.DMChatView.TweetTrove}

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
			app.buffered_render_htmx(w, "dm-composer", global_data, chat_view_data) // Wipe the chat box
			go app.Profile.SaveDMTrove(trove, true)
		}

		chat_view_data.ActiveRoomID = room_id
		chat_view_data.LatestPollingTimestamp = -1
		if latest_timestamp_str := r.URL.Query().Get("latest_timestamp"); latest_timestamp_str != "" {
			var err error
			chat_view_data.LatestPollingTimestamp, err = strconv.Atoi(latest_timestamp_str)
			panic_if(err)
		}
		if r.URL.Query().Get("scroll_bottom") != "0" {
			chat_view_data.ScrollBottom = true
		}
		chat_contents := app.Profile.GetChatRoomContents(room_id, chat_view_data.LatestPollingTimestamp)
		chat_view_data.MergeWith(chat_contents.DMTrove)
		chat_view_data.MessageIDs = chat_contents.MessageIDs
		if len(chat_view_data.MessageIDs) > 0 {
			last_message_id := chat_view_data.MessageIDs[len(chat_view_data.MessageIDs)-1]
			chat_view_data.LatestPollingTimestamp = int(chat_view_data.Messages[last_message_id].SentAt.UnixMilli())
		}

		if r.URL.Query().Has("poll") || len(parts) == 2 && parts[1] == "send" {
			app.buffered_render_htmx(w, "messages-with-poller", global_data, chat_view_data)
			return
		}
	}

	app.buffered_render_page(w, "tpl/messages.tpl", global_data, chat_view_data)
}
