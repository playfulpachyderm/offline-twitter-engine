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

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type ListData struct {
	List
	Feed      Feed
	UserIDs   []UserID
	ActiveTab string
}

func NewListData(users []User) (ListData, TweetTrove) {
	trove := NewTweetTrove()
	data := ListData{
		UserIDs: []UserID{},
	}
	for _, u := range users {
		trove.Users[u.ID] = u
		data.UserIDs = append(data.UserIDs, u.ID)
	}
	return data, trove
}

func (app *Application) ListDetailFeed(w http.ResponseWriter, r *http.Request) {
	list := get_list_from_context(r.Context())

	c := NewListCursor(list.ID)
	err := parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, r, "invalid cursor (must be a number)")
		return
	}
	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, ErrEndOfFeed) {
		panic(err)
	}
	if is_htmx(r) && c.CursorPosition == CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
	} else {
		app.buffered_render_page2(
			w,
			"tpl/list.tpl",
			PageGlobalData{Title: list.Name, TweetTrove: feed.TweetTrove},
			ListData{Feed: feed, List: list, ActiveTab: "feed"},
		)
	}
}

func (app *Application) ListDetailUsers(w http.ResponseWriter, r *http.Request) {
	list := get_list_from_context(r.Context())
	users := app.Profile.GetListUsers(list.ID)

	data, trove := NewListData(users)
	data.List = list
	data.ActiveTab = "users"
	app.buffered_render_page2(w, "tpl/list.tpl", PageGlobalData{Title: list.Name, TweetTrove: trove}, data)
}

func (app *Application) ListDelete(w http.ResponseWriter, r *http.Request) {
	list := get_list_from_context(r.Context())
	app.Profile.DeleteList(list.ID)
	http.Redirect(w, r, "/lists", 302)
}

func (app *Application) ListDetail(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'ListDetail' handler (path: %q)", r.URL.Path)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) == 1 && parts[0] == "" {
		switch r.Method {
		case "DELETE":
			app.ListDelete(w, r)
		default:
			// No further path; just show the feed
			app.ListDetailFeed(w, r)
		}
		return
	}

	switch parts[0] {
	case "users":
		app.ListDetailUsers(w, r)
	case "add_user":
		app.ListAddUser(w, r)
	case "remove_user":
		app.ListRemoveUser(w, r)
	default:
		app.error_404(w, r)
	}
}

func (app *Application) ListAddUser(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("user_handle")
	if handle[0] == '@' {
		handle = handle[1:]
	}
	user, err := app.Profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		app.error_400_with_message(w, r, "Fetch user: "+err.Error())
		return
	}
	list := get_list_from_context(r.Context())
	app.Profile.SaveListUser(list.ID, user.ID)
	http.Redirect(w, r, fmt.Sprintf("/lists/%d/users", list.ID), 302)
}

func (app *Application) ListRemoveUser(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("user_handle")
	if handle[0] == '@' {
		handle = handle[1:]
	}
	user, err := app.Profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		app.error_400_with_message(w, r, "Fetch user: "+err.Error())
		return
	}
	list := get_list_from_context(r.Context())
	app.Profile.DeleteListUser(list.ID, user.ID)
	http.Redirect(w, r, fmt.Sprintf("/lists/%d/users", list.ID), 302)
}

func (app *Application) Lists(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'Lists' handler (path: %q)", r.URL.Path)

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// List detail
	if parts[0] != "" { // If there's an ID param
		_list_id, err := strconv.Atoi(parts[0])
		if err != nil {
			app.error_400_with_message(w, r, "List ID must be a number")
			return
		}
		list, err := app.Profile.GetListById(ListID(_list_id))
		if err != nil {
			app.error_404(w, r)
			return
		}
		req_with_ctx := r.WithContext(add_list_to_context(r.Context(), list))
		http.StripPrefix(fmt.Sprintf("/%d", list.ID), http.HandlerFunc(app.ListDetail)).ServeHTTP(w, req_with_ctx)
		return
	}

	// New list
	if r.Method == "POST" {
		var formdata struct {
			Name string `json:"name"`
		}
		data, err := io.ReadAll(r.Body)
		panic_if(err)
		err = json.Unmarshal(data, &formdata)
		panic_if(err)
		new_list := List{Name: formdata.Name}
		app.Profile.SaveList(&new_list)
		http.Redirect(w, r, fmt.Sprintf("/lists/%d/users", new_list.ID), 302)
		return
	}

	// List index
	lists := app.Profile.GetAllLists()
	trove := NewTweetTrove()
	for _, l := range lists {
		for _, u := range l.Users {
			trove.Users[u.ID] = u
		}
	}
	app.buffered_render_page2(
		w,
		"tpl/list_of_lists.tpl",
		PageGlobalData{Title: "Lists", TweetTrove: trove},
		lists,
	)
}

const LIST_KEY = key("list") // type `key` is defined in "handler_tweet_detail"

func add_list_to_context(ctx context.Context, list List) context.Context {
	return context.WithValue(ctx, LIST_KEY, list)
}

func get_list_from_context(ctx context.Context) List {
	list, is_ok := ctx.Value(LIST_KEY).(List)
	if !is_ok {
		panic("List not found in context")
	}
	return list
}
