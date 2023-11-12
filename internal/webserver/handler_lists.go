package webserver

import (
	"net/http"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (app *Application) Lists(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Lists' handler (path: %q)", r.URL.Path)

	var users []scraper.User
	err := app.Profile.DB.Select(&users, `
		select id, display_name, handle, bio, following_count, followers_count, location, website, join_date, is_private, is_verified,
	           is_banned, is_deleted, profile_image_url, profile_image_local_path, banner_image_url, banner_image_local_path,
	           pinned_tweet_id, is_content_downloaded, is_followed
	      from users
	     where is_followed = 1`)
	panic_if(err)

	app.buffered_render_basic_page(w, "tpl/list.tpl", users)
}
