package webserver

import (
	"errors"
	"net/http"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func (app *Application) Timeline(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Timeline' handler (path: %q)", r.URL.Path)

	c := persistence.NewTimelineCursor()
	err := parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, "invalid cursor (must be a number)")
		return
	}

	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEndOfFeed) {
			// TODO
		} else {
			panic(err)
		}
	}

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
	} else {
		app.buffered_render_page(w, "tpl/offline_timeline.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
	}
}
