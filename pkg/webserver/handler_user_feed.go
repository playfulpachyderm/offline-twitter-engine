package webserver

import (
	"errors"
	"net/http"
	"strings"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (app *Application) UserFeed(w http.ResponseWriter, r *http.Request) {
	app.TraceLog.Printf("'UserFeed' handler (path: %q)", r.URL.Path)

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	user, err := app.Profile.GetUserByHandle(UserHandle(parts[0]))

	if errors.Is(err, ErrNotInDatabase) {
		if !app.IsScrapingDisabled {
			user, err = app.API.GetUser(UserHandle(parts[0]))
		}
		if err != nil { // ErrDoesntExist or otherwise
			app.error_404(w, r)
			return
		}
		panic_if(app.Profile.SaveUser(&user)) // TODO: handle conflicting users
		panic_if(app.Profile.DownloadUserContentFor(&user, app.API.DownloadMedia))
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
	if len(parts) > 1 && parts[1] == "followers_you_know" {
		app.UserFollowersYouKnow(w, r, user)
		return
	}
	if len(parts) > 1 && parts[1] == "followees_you_know" {
		app.UserFolloweesYouKnow(w, r, user)
		return
	}
	if len(parts) > 1 && parts[1] == "mutual_followers" {
		app.UserMutualFollowers(w, r, user)
		return
	}

	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		// Update the user themself
		user, err = app.API.GetUser(UserHandle(parts[0]))
		panic_if(err)

		// Get followers-you-know, while we're at it
		trove, err := app.API.GetFollowersYouKnow(user.ID, 200) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			panic(err) // Let 500 handler catch it
		}
		app.full_save_tweet_trove(trove)
		app.Profile.SaveAsFolloweesList(app.ActiveUser.ID, trove) // You follow everyone in this list...
		app.Profile.SaveAsFollowersList(user.ID, trove)           // ...and they follow the target user

		panic_if(app.Profile.SaveUser(&user)) // TODO: handle conflicting users
		panic_if(app.Profile.DownloadUserContentFor(&user, app.API.DownloadMedia))

		if len(parts) == 1 { // The URL is just the user handle
			// Run scraper
			trove, err := app.API.GetUserFeed(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.full_save_tweet_trove(trove)
		} else if len(parts) == 2 && parts[1] == "likes" {
			trove, err := app.API.GetUserLikes(user.ID, 50) // TODO: parameterizable
			if err != nil {
				app.ErrorLog.Print(err)
				// TOOD: show error in UI
			}
			app.full_save_tweet_trove(trove)
		}
	}

	// Add more stuff to the user (this has to be done after scraping or else it will get clobbered)
	user.IsFollowed = app.Profile.IsXFollowingY(app.ActiveUser.ID, user.ID)
	user.IsFollowingYou = app.Profile.IsXFollowingY(user.ID, app.ActiveUser.ID)
	user.Lists = app.Profile.GetListsForUser(user.ID)
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	var c Cursor
	if len(parts) > 1 && parts[1] == "likes" {
		c = NewUserFeedLikesCursor(user.Handle)
	} else {
		c = NewUserFeedCursor(user.Handle)
	}
	if len(parts) > 1 && parts[1] == "without_replies" {
		c.FilterReplies = EXCLUDE
	}
	if len(parts) > 1 && parts[1] == "media" {
		c.FilterMedia = REQUIRE
	}
	err = parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, r, "invalid cursor (must be a number)")
		return
	}

	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, ErrEndOfFeed) {
		panic(err)
	}
	feed.Users[user.ID] = user

	data := struct {
		Feed
		UserID
		PinnedTweet Tweet
		FeedType    string
	}{Feed: feed, UserID: user.ID}

	if len(parts) == 2 {
		data.FeedType = parts[1]
	} else {
		data.FeedType = ""
	}

	// Add a pinned tweet if there is one and it's in the DB; otherwise skip
	// Also, only show pinned tweets on default tab (tweets+replies) or "without_replies" tab
	if user.PinnedTweetID != TweetID(0) && (len(parts) <= 1 || parts[1] == "without_replies") {
		data.PinnedTweet, err = app.Profile.GetTweetById(user.PinnedTweetID)
		if err != nil && !errors.Is(err, ErrNotInDatabase) {
			panic(err)
		}
		feed.TweetTrove.Tweets[data.PinnedTweet.ID] = data.PinnedTweet

		// Fetch quoted tweet if necessary
		if data.PinnedTweet.QuotedTweetID != TweetID(0) {
			feed.TweetTrove.Tweets[data.PinnedTweet.QuotedTweetID], err = app.Profile.GetTweetById(data.PinnedTweet.QuotedTweetID)
			if err != nil && !errors.Is(err, ErrNotInDatabase) {
				panic(err)
			}
			// And the user
			qt_user_id := feed.TweetTrove.Tweets[data.PinnedTweet.QuotedTweetID].UserID
			feed.TweetTrove.Users[qt_user_id], err = app.Profile.GetUserByID(qt_user_id)
			panic_if(err)
		}
	}

	if is_htmx(r) && c.CursorPosition == CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	} else {
		app.buffered_render_page(w, "tpl/user_feed.tpl", PageGlobalData{TweetTrove: feed.TweetTrove}, data)
	}
}

type FollowsData struct {
	Title        string
	HeaderUserID UserID
	UserIDs      []UserID
}

func NewFollowsData(users []User) (FollowsData, TweetTrove) {
	trove := NewTweetTrove()
	data := FollowsData{
		UserIDs: []UserID{},
	}
	for _, u := range users {
		trove.Users[u.ID] = u
		data.UserIDs = append(data.UserIDs, u.ID)
	}
	return data, trove
}

func (app *Application) UserFollowees(w http.ResponseWriter, r *http.Request, user User) {
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		// Run scraper
		trove, err := app.API.GetFollowees(user.ID, 200) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			panic(err) // Let 500 handler catch it
		}
		app.full_save_tweet_trove(trove)
		app.Profile.SaveAsFolloweesList(user.ID, trove)
	}
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	data, trove := NewFollowsData(app.Profile.GetFollowees(user.ID))
	trove.Users[user.ID] = user // Not loaded otherwise; needed to profile image in the login button on the sidebar
	data.Title = "Followees"
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}

func (app *Application) UserFollowers(w http.ResponseWriter, r *http.Request, user User) {
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		// Run scraper
		trove, err := app.API.GetFollowers(user.ID, 200) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			panic(err) // Let 500 handler catch it
		}
		app.full_save_tweet_trove(trove)
		app.Profile.SaveAsFollowersList(user.ID, trove)
	}
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	data, trove := NewFollowsData(app.Profile.GetFollowers(user.ID))
	trove.Users[user.ID] = user
	data.Title = "Followers"
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}

func (app *Application) UserFollowersYouKnow(w http.ResponseWriter, r *http.Request, user User) {
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		trove, err := app.API.GetFollowersYouKnow(user.ID, 200) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			panic(err) // Let 500 handler catch it
		}
		app.full_save_tweet_trove(trove)
		app.Profile.SaveAsFolloweesList(app.ActiveUser.ID, trove) // You follow everyone in this list...
		app.Profile.SaveAsFollowersList(user.ID, trove)           // ...and they follow the target user
	}
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	data, trove := NewFollowsData(app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID))
	trove.Users[user.ID] = user
	data.Title = "Followers you know"
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}

func (app *Application) UserFolloweesYouKnow(w http.ResponseWriter, r *http.Request, user User) {
	if r.URL.Query().Has("scrape") {
		app.error_400_with_message(w, r, "This page can't be scraped (it's Offline Twitter only)")
	}
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	data, trove := NewFollowsData(app.Profile.GetFolloweesYouKnow(app.ActiveUser.ID, user.ID))
	trove.Users[user.ID] = user
	data.Title = "Followees you know"
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}

func (app *Application) UserMutualFollowers(w http.ResponseWriter, r *http.Request, user User) {
	if r.URL.Query().Has("scrape") {
		app.error_400_with_message(w, r, "This page can't be scraped (it's Offline Twitter only)")
	}
	user.FollowersYouKnow = app.Profile.GetFollowersYouKnow(app.ActiveUser.ID, user.ID)

	data, trove := NewFollowsData(app.Profile.GetMutualFollowers(user.ID))
	trove.Users[user.ID] = user
	data.Title = "Mutual followers"
	data.HeaderUserID = user.ID
	app.buffered_render_page(w, "tpl/follows.tpl", PageGlobalData{TweetTrove: trove}, data)
}
