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
	if errors.Is(err, persistence.ErrNotInDatabase) {
		if !app.IsScrapingDisabled {
			user, err = app.API.GetUser(scraper.UserHandle(parts[0]))
		}
		if err != nil { // ErrDoesntExist or otherwise
			app.error_404(w, r)
			return
		}
		panic_if(app.Profile.SaveUser(&user)) // TODO: handle conflicting users
		panic_if(app.Profile.DownloadUserContentFor(&user, &app.API))
	} else if err != nil {
		panic(err)
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
			trove, err := app.API.GetUserFeed(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.Profile.SaveTweetTrove(trove, false, &app.API)
			go app.Profile.SaveTweetTrove(trove, true, &app.API)
		} else if len(parts) == 2 && parts[1] == "likes" {
			trove, err := app.API.GetUserLikes(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.Profile.SaveTweetTrove(trove, false, &app.API)
			go app.Profile.SaveTweetTrove(trove, true, &app.API)
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
		app.error_400_with_message(w, r, "invalid cursor (must be a number)")
		return
	}

	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, persistence.ErrEndOfFeed) {
		panic(err)
	}
	feed.Users[user.ID] = user

	data := struct {
		persistence.Feed
		scraper.UserID
		PinnedTweet scraper.Tweet
		FeedType    string
	}{Feed: feed, UserID: user.ID}

	if len(parts) == 2 {
		data.FeedType = parts[1]
	} else {
		data.FeedType = ""
	}

	// Add a pinned tweet if there is one and it's in the DB; otherwise skip
	// Also, only show pinned tweets on default tab (tweets+replies) or "without_replies" tab
	if user.PinnedTweetID != scraper.TweetID(0) && (len(parts) <= 1 || parts[1] == "without_replies") {
		data.PinnedTweet, err = app.Profile.GetTweetById(user.PinnedTweetID)
		if err != nil && !errors.Is(err, persistence.ErrNotInDatabase) {
			panic(err)
		}
		feed.TweetTrove.Tweets[data.PinnedTweet.ID] = data.PinnedTweet

		// Fetch quoted tweet if necessary
		if data.PinnedTweet.QuotedTweetID != scraper.TweetID(0) {
			feed.TweetTrove.Tweets[data.PinnedTweet.QuotedTweetID], err = app.Profile.GetTweetById(data.PinnedTweet.QuotedTweetID)
			if err != nil && !errors.Is(err, persistence.ErrNotInDatabase) {
				panic(err)
			}
			// And the user
			qt_user_id := feed.TweetTrove.Tweets[data.PinnedTweet.QuotedTweetID].UserID
			feed.TweetTrove.Users[qt_user_id], err = app.Profile.GetUserByID(qt_user_id)
			panic_if(err)
		}
	}

	if is_htmx(r) && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	} else {
		app.buffered_render_page(w, "tpl/user_feed.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	}
}

type FollowsData struct {
	Title        string
	HeaderUserID scraper.UserID
	UserIDs      []scraper.UserID
}

func NewFollowsData(users []scraper.User) (FollowsData, scraper.TweetTrove) {
	trove := scraper.NewTweetTrove()
	data := FollowsData{
		UserIDs: []scraper.UserID{},
	}
	for _, u := range users {
		trove.Users[u.ID] = u
		data.UserIDs = append(data.UserIDs, u.ID)
	}
	return data, trove
}

func (app *Application) UserFollowees(w http.ResponseWriter, r *http.Request, user scraper.User) {
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			http.Error(w, "Scraping is disabled (are you logged in?)", 401)
			return
		}

		// Run scraper
		trove, err := app.API.GetFollowees(user.ID, 200) // TODO: parameterizable
		if err != nil {
			app.ErrorLog.Print(err)
			// TOOD: show error in UI
		}
		app.Profile.SaveTweetTrove(trove, false, &app.API)
		app.Profile.SaveAsFolloweesList(user.ID, trove)
		go app.Profile.SaveTweetTrove(trove, true, &app.API)
	}

	data, trove := NewFollowsData(app.Profile.GetFollowees(user.ID))
	trove.Users[user.ID] = user // Not loaded otherwise; needed to profile image in the login button on the sidebar
	data.Title = fmt.Sprintf("Followed by @%s", user.Handle)
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}

func (app *Application) UserFollowers(w http.ResponseWriter, r *http.Request, user scraper.User) {
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			http.Error(w, "Scraping is disabled (are you logged in?)", 401)
			return
		}

		// Run scraper
		trove, err := app.API.GetFollowers(user.ID, 200) // TODO: parameterizable
		if err != nil {
			app.ErrorLog.Print(err)
			// TOOD: show error in UI
		}
		app.Profile.SaveTweetTrove(trove, false, &app.API)
		app.Profile.SaveAsFollowersList(user.ID, trove)
		go app.Profile.SaveTweetTrove(trove, true, &app.API)
	}

	data, trove := NewFollowsData(app.Profile.GetFollowers(user.ID))
	trove.Users[user.ID] = user
	data.Title = fmt.Sprintf("@%s's followers", user.Handle)
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}
