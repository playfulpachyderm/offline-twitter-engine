package webserver

import (
	"net/http"
	"strings"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/tracing"
)

func (app *Application) UserFollow(w http.ResponseWriter, r *http.Request) {
	_span := tracing.GetActiveSpan(r.Context()).AddChild("follow_user")
	defer _span.End()
	app.TraceLog.Printf("'UserFollow' handler (path: %q)", r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		app.error_400_with_message(w, r, "Bad URL: "+r.URL.Path)
		return
	}
	span := tracing.GetActiveSpan(r.Context()).AddChild("get_user_by_handle")
	user, err := app.Profile.GetUserByHandle(UserHandle(parts[1]))
	if err != nil {
		app.error_404(w, r)
		return
	}
	span.End()

	panic_if(app.API.FollowUser(user.ID))
	app.Profile.SaveFollow(app.ActiveUser.ID, user.ID)
	user.IsFollowed = true

	app.buffered_render_htmx2(w, r, "following-button", PageGlobalData{}, user)
}

func (app *Application) UserUnfollow(w http.ResponseWriter, r *http.Request) {
	_span := tracing.GetActiveSpan(r.Context()).AddChild("unfollow_user")
	defer _span.End()
	app.TraceLog.Printf("'UserUnfollow' handler (path: %q)", r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		app.error_400_with_message(w, r, "Bad URL: "+r.URL.Path)
		return
	}
	span := tracing.GetActiveSpan(r.Context()).AddChild("get_user_by_handle")
	user, err := app.Profile.GetUserByHandle(UserHandle(parts[1]))
	if err != nil {
		app.error_404(w, r)
		return
	}
	span.End()

	panic_if(app.API.UnfollowUser(user.ID))
	app.Profile.DeleteFollow(app.ActiveUser.ID, user.ID)
	user.IsFollowed = false

	app.buffered_render_htmx2(w, r, "following-button", PageGlobalData{}, user)
}
