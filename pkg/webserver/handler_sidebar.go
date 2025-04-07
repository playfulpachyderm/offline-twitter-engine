package webserver

import (
	"net/http"
)

func (app *Application) NavSidebarPollUpdates(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'NavSidebarPollUpdates' handler (path: %q)", r.URL.Path)

	// Must be an HTMX request, otherwise HTTP 400
	if !is_htmx(r) {
		app.error_400_with_message(w, r, "This is an HTMX-only endpoint, not a page")
		return
	}

	data := NotificationBubbles{
		NumMessageNotifications: len(app.Profile.GetUnreadConversations(app.ActiveUser.ID)),
	}
	if app.LastReadNotificationSortIndex != 0 {
		data.NumRegularNotifications = app.Profile.GetUnreadNotificationsCount(app.ActiveUser.ID, app.LastReadNotificationSortIndex)
	}
	app.buffered_render_htmx(w, "nav-sidebar", PageGlobalData{}, data)
}
