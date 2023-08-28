package webserver

import (
	"errors"
	"net/http"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type UserProfileData struct {
	persistence.Feed
	scraper.UserID
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
		app.error_404(w)
		return
	}

	if len(parts) == 2 && parts[1] == "scrape" {
		if app.IsScrapingDisabled {
			http.Error(w, "Scraping is disabled (are you logged in?)", 401)
			return
		}

		// Run scraper
		trove, err := scraper.GetUserFeedGraphqlFor(user.ID, 50) // TODO: parameterizable
		if err != nil {
			app.ErrorLog.Print(err)
			// TOOD: show error in UI
		}
		app.Profile.SaveTweetTrove(trove)
	}

	c := persistence.NewUserFeedCursor(user.Handle)
	err = parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, "invalid cursor (must be a number)")
		return
	}

	feed, err := app.Profile.NextPage(c)
	if err != nil {
		if errors.Is(err, persistence.ErrEndOfFeed) {
			// TODO
		} else {
			panic(err)
		}
	}
	feed.Users[user.ID] = user

	data := UserProfileData{Feed: feed, UserID: user.ID}

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_tweet_htmx(w, "timeline", data)
	} else {
		app.buffered_render_tweet_page(w, "tpl/user_feed.tpl", data)
	}
}
