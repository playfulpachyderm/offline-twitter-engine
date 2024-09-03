package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/Masterminds/sprig/v3"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type NotificationBubbles struct {
	NumMessageNotifications int
	NumRegularNotifications int
}

// TODO: this name sucks
type PageGlobalData struct {
	scraper.TweetTrove
	SearchText     string
	FocusedTweetID scraper.TweetID
	Toasts         []Toast
	NotificationBubbles
}

func (d PageGlobalData) Tweet(id scraper.TweetID) scraper.Tweet {
	return d.Tweets[id]
}
func (d PageGlobalData) User(id scraper.UserID) scraper.User {
	return d.Users[id]
}
func (d PageGlobalData) Retweet(id scraper.TweetID) scraper.Retweet {
	return d.Retweets[id]
}
func (d PageGlobalData) Space(id scraper.SpaceID) scraper.Space {
	return d.Spaces[id]
}
func (d PageGlobalData) Notification(id scraper.NotificationID) scraper.Notification {
	return d.Notifications[id]
}
func (d PageGlobalData) Message(id scraper.DMMessageID) scraper.DMMessage {
	return d.Messages[id]
}
func (d PageGlobalData) ChatRoom(id scraper.DMChatRoomID) scraper.DMChatRoom {
	return d.Rooms[id]
}
func (d PageGlobalData) GetFocusedTweetID() scraper.TweetID {
	return d.FocusedTweetID
}
func (d PageGlobalData) GetSearchText() string {
	fmt.Println(d.SearchText)
	return d.SearchText
}
func (d PageGlobalData) GlobalData() PageGlobalData {
	return d
}

// Config object for buffered rendering
type renderer struct {
	Funcs     template.FuncMap
	Filenames []string
	TplName   string
	Data      interface{}
}

// Render the given template using a bytes.Buffer.  This avoids the possibility of failing partway
// through the rendering, and sending an imcomplete response with "Bad Request" or "Server Error" at the end.
func (r renderer) BufferedRender(w io.Writer) {
	var tpl *template.Template
	var err error

	funcs := sprig.FuncMap()
	for i := range r.Funcs {
		funcs[i] = r.Funcs[i]
	}
	if use_embedded == "true" {
		tpl, err = template.New("").Funcs(funcs).ParseFS(embedded_files, r.Filenames...)
	} else {
		tpl, err = template.New("").Funcs(funcs).ParseFiles(r.Filenames...)
	}
	panic_if(err)

	buf := new(bytes.Buffer)
	err = tpl.ExecuteTemplate(buf, r.TplName, r.Data)
	panic_if(err)

	_, err = buf.WriteTo(w)
	panic_if(err)
}

// Render the "base" template, creating a full HTML page corresponding to the given template file,
// with all available partials.
func (app *Application) buffered_render_page(w http.ResponseWriter, tpl_file string, global_data PageGlobalData, tpl_data interface{}) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	global_data.NotificationBubbles.NumMessageNotifications = len(app.Profile.GetUnreadConversations(app.ActiveUser.ID))
	if app.LastReadNotificationSortIndex != 0 {
		global_data.NotificationBubbles.NumRegularNotifications = app.Profile.GetUnreadNotificationsCount(
			app.ActiveUser.ID,
			app.LastReadNotificationSortIndex,
		)
	}

	r := renderer{
		Funcs:     app.make_funcmap(global_data),
		Filenames: append(partials, get_filepath(tpl_file)),
		TplName:   "base",
		Data:      tpl_data,
	}
	r.BufferedRender(w)
}

// Render a particular template (HTMX response, i.e., not a full page)
func (app *Application) buffered_render_htmx(w http.ResponseWriter, tpl_name string, global_data PageGlobalData, tpl_data interface{}) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	r := renderer{
		Funcs:     app.make_funcmap(global_data),
		Filenames: partials,
		TplName:   tpl_name,
		Data:      tpl_data,
	}
	r.BufferedRender(w)
}

// Assemble the list of funcs that can be used in the templates
func (app *Application) make_funcmap(global_data PageGlobalData) template.FuncMap {
	return template.FuncMap{
		// Get data from the global objects
		"tweet":            global_data.Tweet,
		"user":             global_data.User,
		"retweet":          global_data.Retweet,
		"space":            global_data.Space,
		"notification":     global_data.Notification,
		"dm_message":       global_data.Message,
		"chat_room":        global_data.ChatRoom,
		"focused_tweet_id": global_data.GetFocusedTweetID,
		"search_text":      global_data.GetSearchText,
		"global_data":      global_data.GlobalData, // This fucking sucks
		"active_user": func() scraper.User {
			return app.ActiveUser
		},

		// Utility functions
		"get_tombstone_text": func(t scraper.Tweet) string {
			if t.TombstoneText != "" {
				return t.TombstoneText
			}
			return t.TombstoneType
		},
		"cursor_to_query_params": func(c persistence.Cursor) string {
			result := url.Values{}
			result.Set("cursor", fmt.Sprint(c.CursorValue))
			result.Set("sort-order", c.SortOrder.String())
			return result.Encode()
		},
		"get_entities": get_entities,
	}
}

type EntityType int

const (
	ENTITY_TYPE_TEXT EntityType = iota
	ENTITY_TYPE_MENTION
	ENTITY_TYPE_HASHTAG
)

type Entity struct {
	EntityType
	Contents string
}

func get_entities(text string) []Entity {
	ret := []Entity{}
	start := 0
	for _, idxs := range regexp.MustCompile(`(\W|^)[@#]\w+`).FindAllStringIndex(text, -1) {
		// The character immediately preceding the entity must not be a word character (alphanumeric
		// or "_").  This is to avoid matching emails.  Accordingly, if the first character in the
		// match isn't a '@' or '#' (i.e., there's a preceding character), skip past it.
		if text[idxs[0]] != '@' && text[idxs[0]] != '#' {
			idxs[0] += 1
		}
		if start != idxs[0] {
			ret = append(ret, Entity{ENTITY_TYPE_TEXT, text[start:idxs[0]]})
		}
		piece := text[idxs[0]+1 : idxs[1]] // Chop off the "#" or "@"
		if text[idxs[0]] == '@' {
			ret = append(ret, Entity{ENTITY_TYPE_MENTION, piece})
		} else {
			ret = append(ret, Entity{ENTITY_TYPE_HASHTAG, piece})
		}
		start = idxs[1]
	}
	if start < len(text) {
		ret = append(ret, Entity{ENTITY_TYPE_TEXT, text[start:]})
	}

	return ret
}

func is_htmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
