package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (app *Application) UserFeed(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'UserFeed' handler (path: %q)", r.URL.Path)

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	user, err := app.Profile.GetUserByHandle(scraper.UserHandle(parts[0]))
	if err != nil {
		if !app.IsScrapingDisabled {
			user, err = scraper.GetUser(scraper.UserHandle(parts[0]))
		}
		if err != nil {
			app.error_404(w)
			return
		}
		panic_if(app.Profile.SaveUser(&user))
		panic_if(app.Profile.DownloadUserContentFor(&user))
	}

	if len(parts) > 1 && parts[1] == "followers" {
		app.UserFollowers(w, r, user)
		return
	}
	if len(parts) > 1 && parts[1] == "followees" {
		app.UserFollowees(w, r, user)
		return
	}

	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			http.Error(w, "Scraping is disabled (are you logged in?)", 401)
			return
		}

		if len(parts) == 1 { // The URL is just the user handle
			// Run scraper
			trove, err := scraper.GetUserFeedGraphqlFor(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.Profile.SaveTweetTrove(trove, false)
			go app.Profile.SaveTweetTrove(trove, true)
		} else if len(parts) == 2 && parts[1] == "likes" {
			trove, err := scraper.GetUserLikes(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.Profile.SaveTweetTrove(trove, false)
			go app.Profile.SaveTweetTrove(trove, true)
		}
	}

	var c persistence.Cursor
	if len(parts) > 1 && parts[1] == "likes" {
		c = persistence.NewUserFeedLikesCursor(user.Handle)
	} else {
		c = persistence.NewUserFeedCursor(user.Handle)
	}
	if len(parts) > 1 && parts[1] == "without_replies" {
		c.FilterReplies = persistence.EXCLUDE
	}
	if len(parts) > 1 && parts[1] == "media" {
		c.FilterMedia = persistence.REQUIRE
	}
	err = parse_cursor_value(&c, r)
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
	feed.Users[user.ID] = user

	data := struct {
		persistence.Feed
		scraper.UserID
		FeedType string
	}{Feed: feed, UserID: user.ID}

	if len(parts) == 2 {
		data.FeedType = parts[1]
	} else {
		data.FeedType = ""
	}

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	} else {
		app.buffered_render_page(w, "tpl/user_feed.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	}
}

func (app *Application) UserFollowees(w http.ResponseWriter, r *http.Request, user scraper.User) {
	data, trove := NewListData(app.Profile.GetFollowees(user.ID))
	trove.Users[user.ID] = user // Not loaded otherwise; needed to profile image in the login button on the sidebar
	data.Title = fmt.Sprintf("Followed by @%s", user.Handle)
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/list.tpl", PageGlobalData{TweetTrove: trove}, data)
}
func (app *Application) UserFollowers(w http.ResponseWriter, r *http.Request, user scraper.User) {
	data, trove := NewListData(app.Profile.GetFollowers(user.ID))
	trove.Users[user.ID] = user
	data.Title = fmt.Sprintf("@%s's followers", user.Handle)
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/list.tpl", PageGlobalData{TweetTrove: trove}, data)
}
