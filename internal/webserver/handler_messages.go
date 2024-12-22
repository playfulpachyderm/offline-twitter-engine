package webserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (app *Application) message_mark_as_read(w http.ResponseWriter, r *http.Request) {
	room_id := get_room_id_from_context(r.Context())

	c := persistence.NewConversationCursor(room_id)
	c.PageSize = 1
	chat_contents := app.Profile.GetChatRoomMessagesByCursor(c)
	last_message_id := chat_contents.MessageIDs[len(chat_contents.MessageIDs)-1]
	if app.IsScrapingDisabled {
		app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
		app.error_401(w, r)
		return
	}
	panic_if(app.API.MarkDMChatRead(room_id, last_message_id))
	room := chat_contents.Rooms[room_id]
	participant, is_ok := room.Participants[app.ActiveUser.ID]
	if !is_ok {
		panic(room)
	}
	participant.LastReadEventID = last_message_id
	room.Participants[app.ActiveUser.ID] = participant
	panic_if(app.Profile.SaveChatRoom(room))
	app.toast(w, r, 200, Toast{
		Title:          "Success",
		Message:        `Conversation marked as "read"`,
		Type:           "success",
		AutoCloseDelay: 2000,
	})
}

func (app *Application) message_send(w http.ResponseWriter, r *http.Request) {
	room_id := get_room_id_from_context(r.Context())

	body, err := io.ReadAll(r.Body)
	panic_if(err)
	var message_data struct {
		Text        string `json:"text"`
		InReplyToID string `json:"in_reply_to_id"`
	}
	panic_if(json.Unmarshal(body, &message_data))

	in_reply_to_id, err := strconv.Atoi(message_data.InReplyToID)
	if err != nil {
		in_reply_to_id = 0
	}
	if app.IsScrapingDisabled {
		app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
		app.error_401(w, r)
		return
	}
	trove, err := app.API.SendDMMessage(room_id, message_data.Text, scraper.DMMessageID(in_reply_to_id))
	if err != nil {
		panic(err)
	}
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)
}

func (app *Application) message_detail(w http.ResponseWriter, r *http.Request) {
	room_id := get_room_id_from_context(r.Context())

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	is_sending := len(parts) == 1 && parts[0] == "send"

	chat_view_data, global_data := app.get_message_global_data()
	if _, is_ok := chat_view_data.Rooms[room_id]; !is_ok {
		app.error_404(w, r)
		return
	}

	if len(parts) == 1 && parts[0] == "mark-as-read" {
		app.message_mark_as_read(w, r)
		return
	}

	// Handle reactions
	if len(parts) == 1 && parts[0] == "reacc" {
		if app.IsScrapingDisabled {
			app.error_401(w, r)
			return
		}
		var data struct {
			MessageID scraper.DMMessageID `json:"message_id,string"`
			Reacc     string              `json:"reacc"`
		}
		data_, err := io.ReadAll(r.Body)
		panic_if(err)
		panic_if(json.Unmarshal(data_, &data))
		panic_if(app.API.SendDMReaction(room_id, data.MessageID, data.Reacc))

		dm_message, is_ok := global_data.Messages[data.MessageID]
		if !is_ok {
			// TODO: this seems kind of silly to use global data in the first place; it's unlikely
			// to have the relevant tweet in it, since it just gets the one latest tweet from the
			// convo.  This handler should be pulled out and just fetch the tweet directly--
			// performance probably doesn't matter, but it's spaghetti code otherwise
			trove_with_dm_message, err := app.Profile.GetChatMessage(data.MessageID)
			panic_if(err)
			global_data.MergeWith(trove_with_dm_message)
			dm_message, is_ok = global_data.Messages[data.MessageID]
			if !is_ok {
				panic(global_data)
			}
		}
		dm_message.Reactions[app.ActiveUser.ID] = scraper.DMReaction{
			ID:          0, // Hopefully will be OK temporarily
			DMMessageID: dm_message.ID,
			SenderID:    app.ActiveUser.ID,
			SentAt:      scraper.Timestamp{time.Now()},
			Emoji:       data.Reacc,
		}
		global_data.Messages[dm_message.ID] = dm_message
		app.buffered_render_htmx(w, "message", global_data, dm_message)
		return
	}

	// First send a message, if applicable
	if is_sending {
		app.message_send(w, r)
	}

	if r.URL.Query().Has("scrape") && !app.IsScrapingDisabled {
		max_id := scraper.DMMessageID(^uint(0) >> 1)
		trove, err := app.API.GetConversation(room_id, max_id, 50) // TODO: parameterizable
		if err != nil {
			panic(err)
		}
		app.Profile.SaveTweetTrove(trove, false, &app.API)
		go app.Profile.SaveTweetTrove(trove, true, &app.API) // Download the content in the background
	}

	// `LatestPollingTimestamp` sort of passes-through the function; if we're not updating it, it
	//  goes out the same it came in.  Hence, using a single variable for it
	chat_view_data.LatestPollingTimestamp = 0
	chat_view_data.ActiveRoomID = room_id
	if latest_timestamp_str := r.URL.Query().Get("latest_timestamp"); latest_timestamp_str != "" {
		var err error
		chat_view_data.LatestPollingTimestamp, err = strconv.Atoi(latest_timestamp_str)
		panic_if(err) // TODO: 400 not 500
	}
	if r.URL.Query().Get("scroll_bottom") != "0" {
		chat_view_data.ScrollBottom = true
	}

	c := persistence.NewConversationCursor(room_id)
	c.SinceTimestamp = scraper.TimestampFromUnixMilli(int64(chat_view_data.LatestPollingTimestamp))
	if cursor_value := r.URL.Query().Get("cursor"); cursor_value != "" {
		until_time, err := strconv.Atoi(cursor_value)
		panic_if(err) // TODO: 400 not 500
		c.UntilTimestamp = scraper.TimestampFromUnixMilli(int64(until_time))
	}
	chat_contents := app.Profile.GetChatRoomMessagesByCursor(c)
	chat_view_data.DMChatView.MergeWith(chat_contents.TweetTrove)
	chat_view_data.MessageIDs = chat_contents.MessageIDs
	chat_view_data.Cursor = chat_contents.Cursor
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

		// Scrolling-back should add new messages to the top of the page
		if r.URL.Query().Has("cursor") {
			app.buffered_render_htmx(w, "messages-top", global_data, chat_view_data)
			return
		}

		// Reload the whole chat view pane
		if r.URL.Query().Has("scrape") {
			app.buffered_render_htmx(w, "chat-view", global_data, chat_view_data)
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

func (app *Application) messages_refresh_list(w http.ResponseWriter, r *http.Request) {
	chat_view_data, global_data := app.get_message_global_data()
	chat_view_data.ActiveRoomID = scraper.DMChatRoomID(r.URL.Query().Get("active-chat"))
	app.buffered_render_htmx(w, "chat-list", global_data, chat_view_data)
}

func (app *Application) Messages(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Messages' handler (path: %q)", r.URL.Path)

	if app.ActiveUser.ID == 0 {
		app.error_401(w, r)
		return
	}

	// Every 3 seconds, message detail page will send request to scrape, with `?poll` set
	if r.URL.Query().Has("poll") && !app.IsScrapingDisabled {
		trove, new_cursor, err := app.API.PollInboxUpdates(inbox_cursor)
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
			panic(err)
		}
		inbox_cursor = new_cursor
		app.Profile.SaveTweetTrove(trove, false, &app.API)
		go app.Profile.SaveTweetTrove(trove, true, &app.API)
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if parts[0] == "refresh-list" {
		app.messages_refresh_list(w, r)
		return
	}
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
