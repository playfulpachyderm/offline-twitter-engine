package webserver

import (
	"net/http"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type ListData struct {
	Title         string
	HeaderUserID  scraper.UserID
	HeaderTweetID scraper.TweetID
	UserIDs       []scraper.UserID
}

func NewListData(users []scraper.User) (ListData, scraper.TweetTrove) {
	trove := scraper.NewTweetTrove()
	data := ListData{
		UserIDs: []scraper.UserID{},
	}
	for _, u := range users {
		trove.Users[u.ID] = u
		data.UserIDs = append(data.UserIDs, u.ID)
	}
	return data, trove
}

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

	data, trove := NewListData(users)
	data.Title = "Offline Follows"
	app.buffered_render_page(w, "tpl/list.tpl", PageGlobalData{TweetTrove: trove}, data)
}
