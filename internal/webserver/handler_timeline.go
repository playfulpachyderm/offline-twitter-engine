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

	data := UserProfileData{Feed: feed} // TODO: wrong struct

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_tweet_htmx(w, "timeline", data)
	} else {
		app.buffered_render_tweet_page(w, "tpl/offline_timeline.tpl", data)
	}
}
