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
	IsUsersSearch    bool
	UserIDs          []scraper.UserID
	// TODO: fill out the search text in the search bar as well (needs modifying the base template)
}

func NewSearchPageData() SearchPageData {
	ret := SearchPageData{SortOrderOptions: []string{}, Feed: persistence.NewFeed()}
	for i := 0; i < 4; i++ { // Don't include "Liked At" option which is #4
		ret.SortOrderOptions = append(ret.SortOrderOptions, persistence.SortOrder(i).String())
	}
	return ret
}

func (app *Application) SearchUsers(w http.ResponseWriter, r *http.Request) {
	ret := NewSearchPageData()
	ret.IsUsersSearch = true
	ret.SearchText = strings.Trim(r.URL.Path, "/")
	ret.UserIDs = []scraper.UserID{}
	for _, u := range app.Profile.SearchUsers(ret.SearchText) {
		ret.TweetTrove.Users[u.ID] = u
		ret.UserIDs = append(ret.UserIDs, u.ID)
	}
	app.buffered_render_page(w, "tpl/search.tpl", PageGlobalData{TweetTrove: ret.Feed.TweetTrove, SearchText: ret.SearchText}, ret)
}

func (app *Application) Search(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Search' handler (path: %q)", r.URL.Path)

	search_text := strings.Trim(r.URL.Path, "/")
	if search_text == "" {
		// Redirect GET param "q" to use a URL param instead
		search_text = r.URL.Query().Get("q")
		if search_text == "" {
			app.error_400_with_message(w, r, "Empty search query")
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/search/%s", url.PathEscape(search_text)), 302)
		return
	}

	// Handle users search
	if r.URL.Query().Get("type") == "users" {
		app.SearchUsers(w, r)
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

	// Actual search
	// Scrape if needed
	if r.URL.Query().Has("scrape") {
		if app.IsScrapingDisabled {
			app.InfoLog.Printf("Would have scraped: %s", r.URL.Path)
			app.error_401(w, r)
			return
		}

		// Run scraper
		trove, err := app.API.Search(search_text, 1) // TODO: parameterizable
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			app.ErrorLog.Print(err)
			// TOOD: show error in UI
		}
		app.Profile.SaveTweetTrove(trove, false, app.API.DownloadMedia)
		go app.Profile.SaveTweetTrove(trove, true, app.API.DownloadMedia)
	}

	c, err := persistence.NewCursorFromSearchQuery(search_text)
	if err != nil {
		app.error_400_with_message(w, r, err.Error())
		return
	}
	err = parse_cursor_value(&c, r)
	if err != nil {
		app.error_400_with_message(w, r, "invalid cursor (must be a number)")
		return
	}
	var is_ok bool
	c.SortOrder, is_ok = persistence.SortOrderFromString(r.URL.Query().Get("sort-order"))
	if !is_ok && r.URL.Query().Get("sort-order") != "" {
		app.error_400_with_message(w, r, "Invalid sort order")
	}

	feed, err := app.Profile.NextPage(c, app.ActiveUser.ID)
	if err != nil && !errors.Is(err, persistence.ErrEndOfFeed) {
		panic(err)
	}

	data := NewSearchPageData()
	data.Feed = feed
	data.SearchText = search_text
	data.SortOrder = c.SortOrder

	if is_htmx(r) && c.CursorPosition == persistence.CURSOR_MIDDLE {
		// It's a Show More request
		app.buffered_render_htmx(w, "timeline", PageGlobalData{TweetTrove: data.Feed.TweetTrove, SearchText: search_text}, data)
	} else {
		app.buffered_render_page(w, "tpl/search.tpl", PageGlobalData{TweetTrove: data.Feed.TweetTrove, SearchText: search_text}, data)
	}
}
