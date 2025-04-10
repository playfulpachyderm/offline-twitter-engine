package webserver

import (
	"errors"
	"net/http"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (app *Application) Bookmarks(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'Bookmarks' handler (path: %q)", r.URL.Path)

	// Run a scrape if needed
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		// Run scraper
		trove, err := app.API.GetBookmarks(300) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			app.ErrorLog.Print(err)
			panic(err) // Return a toast
		}

		app.full_save_tweet_trove(trove)
	}

	c := NewUserFeedBookmarksCursor(app.ActiveUser.Handle)
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
			"tpl/bookmarks.tpl",
			PageGlobalData{Title: "Bookmarks", TweetTrove: feed.TweetTrove},
			TimelineData{Feed: feed},
		)
	}
}
