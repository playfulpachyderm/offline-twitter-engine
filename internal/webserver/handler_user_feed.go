package webserver

import (
	"fmt"
	"errors"
	"net/http"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type UserProfileData struct {
	persistence.Feed
	scraper.UserID
	FeedType string
}

func (t UserProfileData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t UserProfileData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}
func (t UserProfileData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}
func (t UserProfileData) Space(id scraper.SpaceID) scraper.Space {
	return t.Spaces[id]
}
func (t UserProfileData) FocusedTweetID() scraper.TweetID {
	return scraper.TweetID(0)
}

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

	data := UserProfileData{Feed: feed, UserID: user.ID}
	if len(parts) == 2 {
		data.FeedType = parts[1]
	} else {
		data.FeedType = ""
	}

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_tweet_htmx(w, "timeline", data)
	} else {
		app.buffered_render_tweet_page(w, "tpl/user_feed.tpl", data)
	}
}

type ListData struct {
	Title string
	Users []scraper.User
}

func (app *Application) UserFollowees(w http.ResponseWriter, r *http.Request, user scraper.User) {
	app.buffered_render_basic_page(w, "tpl/list.tpl", ListData{
		Title: fmt.Sprintf("Followed by @%s", user.Handle),
		Users: app.Profile.GetFollowees(user.ID),
	})
}
func (app *Application) UserFollowers(w http.ResponseWriter, r *http.Request, user scraper.User) {
	app.buffered_render_basic_page(w, "tpl/list.tpl", ListData{
		Title: fmt.Sprintf("Followers of @%s", user.Handle),
		Users: app.Profile.GetFollowers(user.ID),
	})
}
