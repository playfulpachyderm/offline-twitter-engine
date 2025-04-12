package webserver

import (
	"errors"
	"net/http"
	"strings"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/tracing"
)

type TimelineData struct {
	Feed
	ActiveTab string
}

// TODO: deprecated-offline-follows

func (app *Application) OfflineTimeline(w http.ResponseWriter, r *http.Request) {
	_span := tracing.GetActiveSpan(r.Context()).AddChild("offline_timeline")
	defer _span.End()
	app.TraceLog.Printf("'Timeline' handler (path: %q)", r.URL.Path)

	c := NewTimelineCursor()
	err := parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, r, "invalid cursor (must be a number)")
		return
	}

	span := tracing.GetActiveSpan(r.Context()).AddChild("cursor_next_page")
	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, ErrEndOfFeed) {
		panic(err)
	}
	span.End()

	if is_htmx(r) && c.CursorPosition == CURSOR_MIDDLE {
		// It's a Show More request
		span := tracing.GetActiveSpan(r.Context()).AddChild("buffered_render_htmx")
		app.buffered_render_htmx2(w, r, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
		span.End()
	} else {
		app.buffered_render_page2(
			w, r,
			"tpl/offline_timeline.tpl",
			PageGlobalData{Title: "Timeline", TweetTrove: feed.TweetTrove},
			TimelineData{Feed: feed, ActiveTab: "Offline"},
		)
	}
}

func (app *Application) Timeline(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) > 1 && parts[1] == "offline" {
		app.OfflineTimeline(w, r)
		return
	}

	_span := tracing.GetActiveSpan(r.Context()).AddChild("home_timeline")
	defer _span.End()
	app.TraceLog.Printf("'Timeline' handler (path: %q)", r.URL.Path)

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

	span := tracing.GetActiveSpan(r.Context()).AddChild("cursor_next_page")
	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, ErrEndOfFeed) {
		panic(err)
	}
	span.End()

	if is_htmx(r) && c.CursorPosition == CURSOR_MIDDLE {
		// It's a Show More request
		span := tracing.GetActiveSpan(r.Context()).AddChild("buffered_render_htmx")
		app.buffered_render_htmx2(w, r, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, feed)
		span.End()
	} else {
		app.buffered_render_page2(
			w, r,
			"tpl/offline_timeline.tpl",
			PageGlobalData{Title: "Timeline", TweetTrove: feed.TweetTrove},
			TimelineData{Feed: feed, ActiveTab: "User feed"},
		)
	}
}
