package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func (app *Application) Search(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Search' handler (path: %q)", r.URL.Path)

	search_text := strings.Trim(r.URL.Path, "/")
	if search_text == "" {
		// Redirect GET param "q" to use a URL param instead
		search_text = r.URL.Query().Get("q")
		if search_text == "" {
			app.error_400_with_message(w, "Empty search query")
			return
			// TODO: return an actual page
		}
		http.Redirect(w, r, fmt.Sprintf("/search/%s", url.PathEscape(search_text)), 302)
		return
	}

	// Handle "@username"
	if search_text[0] == '@' {
		http.Redirect(w, r, fmt.Sprintf("/%s", search_text[1:]), 302)
	}

	c, err := persistence.NewCursorFromSearchQuery(search_text)
	if err != nil {
		app.error_400_with_message(w, err.Error())
		return
		// TODO: return actual page
	}
	err = parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, "invalid cursor (must be a number)")
		return
	}

	feed, err := app.Profile.NextPage(c)
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
		app.buffered_render_tweet_page(w, "tpl/search.tpl", data)
	}
}
