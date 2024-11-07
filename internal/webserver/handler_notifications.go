package webserver

import (
	"net/http"
	"strconv"
	"strings"
)

func (app *Application) Notifications(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Notifications' handler (path: %q)", r.URL.Path)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if parts[0] == "mark-all-as-read" {
		app.NotificationsMarkAsRead(w, r)
		return
	}

	cursor_val := 0
	cursor_param := r.URL.Query().Get("cursor")
	if cursor_param != "" {
		var err error
		cursor_val, err = strconv.Atoi(cursor_param)
		if err != nil {
			app.error_400_with_message(w, r, "invalid cursor (must be a number)")
			return
		}
	}

	feed := app.Profile.GetNotificationsForUser(app.ActiveUser.ID, int64(cursor_val), 50) // TODO: parameterizable

	if is_htmx(r) && cursor_val != 0 {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
	} else {
		app.buffered_render_page(w, "tpl/notifications.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
	}
}

func (app *Application) NotificationsMarkAsRead(w http.ResponseWriter, r *http.Request) {
	err := app.API.MarkNotificationsAsRead()
	if err != nil {
		panic(err)
	}
	app.toast(w, r, Toast{
		Title:          "Success",
		Message:        `Notifications marked as "read"`,
		Type:           "success",
		AutoCloseDelay: 2000,
	})
}
