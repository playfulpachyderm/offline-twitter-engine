package webserver

import (
	"errors"
	"net/http"
	"strings"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type TimelineData struct {
	Feed
	ActiveTab string
}

// TODO: deprecated-offline-follows

func (app *Application) OfflineTimeline(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'Timeline' handler (path: %q)", r.URL.Path)

	c := NewTimelineCursor()
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
		app.buffered_render_page(
			w,
			"tpl/offline_timeline.tpl",
			PageGlobalData{TweetTrove: feed.TweetTrove},
			TimelineData{Feed: feed, ActiveTab: "Offline"},
		)
	}
}

func (app *Application) Timeline(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'Timeline' handler (path: %q)", r.URL.Path)

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) > 1 && parts[1] == "offline" {
		app.OfflineTimeline(w, r)
		return
	}

	c := Cursor{
		Keywords:       []string{},
		ToUserHandles:  []UserHandle{},
		SinceTimestamp: TimestampFromUnix(0),
		UntilTimestamp: TimestampFromUnix(0),
		CursorPosition: CURSOR_START,
		CursorValue:    0,
		SortOrder:      SORT_ORDER_NEWEST,
		PageSize:       50,

		FollowedByUserHandle: app.ActiveUser.Handle,
	}
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
		app.buffered_render_page(
			w,
			"tpl/offline_timeline.tpl",
			PageGlobalData{TweetTrove: feed.TweetTrove},
			TimelineData{Feed: feed, ActiveTab: "User feed"},
		)
	}
}
