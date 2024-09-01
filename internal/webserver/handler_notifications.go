package webserver

import (
	"net/http"
)

func (app *Application) Notifications(w http.ResponseWriter, r *http.Request) {
	feed := app.Profile.GetNotificationsForUser(app.ActiveUser.ID, 0)

	app.buffered_render_page(w, "tpl/notifications.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
}
