package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type SearchPageData struct {
	persistence.Feed
	SearchText       string
	SortOrder        persistence.SortOrder
	SortOrderOptions []string
	// TODO: fill out the search text in the search bar as well (needs modifying the base template)
}

func NewSearchPageData() SearchPageData {
	ret := SearchPageData{SortOrderOptions: []string{}}
	for i := 0; i < 4; i++ { // Don't include "Liked At" option which is #4
		ret.SortOrderOptions = append(ret.SortOrderOptions, persistence.SortOrder(i).String())
	}
	return ret
}

func (t SearchPageData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t SearchPageData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}
func (t SearchPageData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}
func (t SearchPageData) Space(id scraper.SpaceID) scraper.Space {
	return t.Spaces[id]
}
func (t SearchPageData) FocusedTweetID() scraper.TweetID {
	return scraper.TweetID(0)
}

func (app *Application) Search(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Search' handler (path: %q)", r.URL.Path)

	search_text := strings.Trim(r.URL.Path, "/")
	if search_text == "" {
		// Redirect GET param "q" to use a URL param instead
		search_text = r.URL.Query().Get("q")
		if search_text == "" {
			app.error_400_with_message(w, "Empty search query")
			return
			// TODO: return an actual page
		}
		http.Redirect(w, r, fmt.Sprintf("/search/%s", url.PathEscape(search_text)), 302)
		return
	}

	// Handle "@username"
	if search_text[0] == '@' {
		http.Redirect(w, r, fmt.Sprintf("/%s", search_text[1:]), 302)
		return
	}

	// Handle pasted URLs
	maybe_url, err := url.Parse(search_text)
	if err == nil && (maybe_url.Host == "twitter.com" || maybe_url.Host == "mobile.twitter.com" || maybe_url.Host == "x.com") {
		// TODO: use scraper.TryParseTweetUrl for this somehow
		// Problem: it currently only supports tweet URLs
		parts := strings.Split(strings.Trim(maybe_url.Path, "/"), "/")

		// Handle tweet links
		if len(parts) == 3 && parts[1] == "status" {
			id, err := strconv.Atoi(parts[2])
			if err == nil {
				http.Redirect(w, r, fmt.Sprintf("/tweet/%d", id), 302)
				return
			}
		}

		// Handle user profile links
		if len(parts) == 1 || (len(parts) == 2 && parts[1] == "with_replies") {
			http.Redirect(w, r, fmt.Sprintf("/%s", parts[0]), 302)
			return
		}
	}

	c, err := persistence.NewCursorFromSearchQuery(search_text)
	if err != nil {
		app.error_400_with_message(w, err.Error())
		return
		// TODO: return actual page
	}
	err = parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, "invalid cursor (must be a number)")
		return
	}
	var is_ok bool
	c.SortOrder, is_ok = persistence.SortOrderFromString(r.URL.Query().Get("sort-order"))
	if !is_ok && r.URL.Query().Get("sort-order") != "" {
		app.error_400_with_message(w, "Invalid sort order")
	}

	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrEndOfFeed) {
			// TODO
		} else {
			panic(err)
		}
	}

	data := NewSearchPageData()
	data.Feed = feed
	data.SearchText = search_text
	data.SortOrder = c.SortOrder

	if r.Header.Get("HX-Request") == "true" && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_tweet_htmx(w, "timeline", data)
	} else {
		app.buffered_render_tweet_page(w, "tpl/search.tpl", data)
	}
}
