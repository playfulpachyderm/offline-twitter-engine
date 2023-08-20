package webserver

import (
	"net/http"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (app *Application) UserFollow(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'UserFollow' handler (path: %q)", r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		app.error_400_with_message(w, "Bad URL: "+r.URL.Path)
		return
	}
	user, err := app.Profile.GetUserByHandle(scraper.UserHandle(parts[1]))
	if err != nil {
		app.error_404(w)
		return
	}

	app.Profile.SetUserFollowed(&user, true)

	app.buffered_render_basic_htmx(w, "following-button", user)
}

func (app *Application) UserUnfollow(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'UserUnfollow' handler (path: %q)", r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		app.error_400_with_message(w, "Bad URL: "+r.URL.Path)
		return
	}
	user, err := app.Profile.GetUserByHandle(scraper.UserHandle(parts[1]))
	if err != nil {
		app.error_404(w)
		return
	}

	app.Profile.SetUserFollowed(&user, false)
	app.buffered_render_basic_htmx(w, "following-button", user)
}
