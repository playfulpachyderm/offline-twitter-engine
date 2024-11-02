package webserver

import (
	"net/http"
	"strconv"
)

func (app *Application) Notifications(w http.ResponseWriter, r *http.Request) {
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
