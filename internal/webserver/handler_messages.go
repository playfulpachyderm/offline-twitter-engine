package webserver

import (
	"context"
	"encoding/json"
	"fmt"
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
	UnreadRoomIDs          map[scraper.DMChatRoomID]bool
}

func (app *Application) messages_index(w http.ResponseWriter, r *http.Request) {
	chat_view_data, global_data := app.get_message_global_data()
	app.buffered_render_page(w, "tpl/messages.tpl", global_data, chat_view_data)
}

func (app *Application) message_detail(w http.ResponseWriter, r *http.Request) {
	room_id := get_room_id_from_context(r.Context())

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	is_sending := len(parts) == 1 && parts[0] == "send"

	chat_view_data, global_data := app.get_message_global_data()

	// First send a message, if applicable
	if is_sending {
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
	chat_view_data.LatestPollingTimestamp = -1 // TODO: why not 0?  If `0` then it won't generate a SQL `where` clause
	if latest_timestamp_str := r.URL.Query().Get("latest_timestamp"); latest_timestamp_str != "" {
		var err error
		chat_view_data.LatestPollingTimestamp, err = strconv.Atoi(latest_timestamp_str)
		panic_if(err) // TODO: 400 not 500
	}
	if r.URL.Query().Get("scroll_bottom") != "0" {
		chat_view_data.ScrollBottom = true
	}
	chat_contents := app.Profile.GetChatRoomContents(room_id, chat_view_data.LatestPollingTimestamp)
	chat_view_data.MergeWith(chat_contents.DMTrove)
	chat_view_data.DMChatView.MergeWith(chat_contents.DMTrove)
	chat_view_data.MessageIDs = chat_contents.MessageIDs
	if len(chat_view_data.MessageIDs) > 0 {
		last_message_id := chat_view_data.MessageIDs[len(chat_view_data.MessageIDs)-1]
		chat_view_data.LatestPollingTimestamp = int(chat_view_data.Messages[last_message_id].SentAt.UnixMilli())
	}

	if is_htmx(r) {
		// Polling for updates and sending a message should add messages at the bottom of the page (newest)
		if r.URL.Query().Has("poll") || is_sending {
			app.buffered_render_htmx(w, "messages-with-poller", global_data, chat_view_data)
			return
		}
	}

	app.buffered_render_page(w, "tpl/messages.tpl", global_data, chat_view_data)
}

func (app *Application) get_message_global_data() (MessageData, PageGlobalData) {
	// Get message list previews
	chat_view_data := MessageData{DMChatView: app.Profile.GetChatRoomsPreview(app.ActiveUser.ID)}
	chat_view_data.UnreadRoomIDs = make(map[scraper.DMChatRoomID]bool)
	for _, id := range app.Profile.GetUnreadConversations(app.ActiveUser.ID) {
		chat_view_data.UnreadRoomIDs[id] = true
	}

	// Initialize the Global Data from the chat list data (last message previews, etc)
	global_data := PageGlobalData{TweetTrove: chat_view_data.DMChatView.TweetTrove}

	return chat_view_data, global_data
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Messages' handler (path: %q)", r.URL.Path)

	if app.ActiveUser.ID == 0 {
		app.error_401(w)
		return
	}

	// Every 3 seconds, message detail page will send request to scrape, with `?poll` set
	if r.URL.Query().Has("poll") {
		// Not run as a goroutine; this call blocks.  It's not actually "background"
		app.background_dm_polling_scrape()
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	room_id := scraper.DMChatRoomID(parts[0])

	// Messages index
	if room_id == "" {
		app.messages_index(w, r)
		return
	}

	// Message detail
	http.StripPrefix(
		fmt.Sprintf("/%s", room_id),
		http.HandlerFunc(app.message_detail),
	).ServeHTTP(w, r.WithContext(add_room_id_to_context(r.Context(), room_id)))
}

const ROOM_ID_KEY = key("room_id") // type `key` is defined in "handler_tweet_detail"

func add_room_id_to_context(ctx context.Context, room_id scraper.DMChatRoomID) context.Context {
	return context.WithValue(ctx, ROOM_ID_KEY, room_id)
}

func get_room_id_from_context(ctx context.Context) scraper.DMChatRoomID {
	room_id, is_ok := ctx.Value(ROOM_ID_KEY).(scraper.DMChatRoomID)
	if !is_ok {
		panic("room_id not found in context")
	}
	return room_id
}
