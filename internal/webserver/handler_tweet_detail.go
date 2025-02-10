package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

var ErrNotFound = errors.New("not found")

type TweetDetailData struct {
	persistence.TweetDetailView
	MainTweetID scraper.TweetID
}

func NewTweetDetailData() TweetDetailData {
	return TweetDetailData{
		TweetDetailView: persistence.NewTweetDetailView(),
	}
}

func (app *Application) ensure_tweet(id scraper.TweetID, is_forced bool, is_conversation_required bool) (scraper.Tweet, error) {
	is_available := false
	is_needing_scrape := is_forced

	// Check if tweet is already in DB
	tweet, err := app.Profile.GetTweetById(id)
	if err != nil {
		if errors.Is(err, persistence.ErrNotInDatabase) {
			is_needing_scrape = true
			is_available = false
		} else {
			panic(err)
		}
	} else {
		is_available = true
		if !tweet.IsConversationScraped {
			is_needing_scrape = true
		}
	}
	if is_available && !is_conversation_required { // TODO: get rid of this, just force the fetch in subsequent handlers if needed
		is_needing_scrape = false
	}

	if is_needing_scrape && !app.IsScrapingDisabled {
		trove, err := app.API.GetTweetFullAPIV2(id, 50) // TODO: parameterizable

		// Save the trove unless there was an unrecoverable error
		if err == nil || errors.Is(err, scraper.END_OF_FEED) || errors.Is(err, scraper.ErrRateLimited) {
			app.full_save_tweet_trove(trove)
			_, is_available = trove.Tweets[id]
		}

		if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
			return scraper.Tweet{}, fmt.Errorf("scraper error: %w", err)
		}
	} else if is_needing_scrape {
		app.InfoLog.Printf("Would have scraped Tweet: %d", id)
	}

	if !is_available {
		return scraper.Tweet{}, ErrNotFound
	}
	return tweet, nil
}

func (app *Application) LikeTweet(w http.ResponseWriter, r *http.Request) {
	tweet := get_tweet_from_context(r.Context())
	like, err := app.API.LikeTweet(tweet.ID)
	// "Already Liked This Tweet" is no big deal-- we can just update the UI as if it succeeded
	if err != nil && !errors.Is(err, scraper.AlreadyLikedThisTweet) {
		// It's a different error
		panic(err)
	}
	err = app.Profile.SaveLike(like)
	panic_if(err)
	tweet.IsLikedByCurrentUser = true

	app.buffered_render_htmx(w, "likes-count", PageGlobalData{}, tweet)
}
func (app *Application) UnlikeTweet(w http.ResponseWriter, r *http.Request) {
	tweet := get_tweet_from_context(r.Context())
	err := app.API.UnlikeTweet(tweet.ID)
	// As above, "Haven't Liked This Tweet" is no big deal-- we can just update the UI as if the request succeeded
	if err != nil && !errors.Is(err, scraper.HaventLikedThisTweet) {
		// It's a different error
		panic(err)
	}
	err = app.Profile.DeleteLike(scraper.Like{UserID: app.ActiveUser.ID, TweetID: tweet.ID})
	panic_if(err)
	tweet.IsLikedByCurrentUser = false

	app.buffered_render_htmx(w, "likes-count", PageGlobalData{}, tweet)
}

func (app *Application) TweetDetail(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'TweetDetail' handler (path: %q)", r.URL.Path)

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	val, err := strconv.Atoi(parts[1])
	if err != nil {
		app.error_400_with_message(w, r, fmt.Sprintf("Invalid tweet ID: %q", parts[1]))
		return
	}
	tweet_id := scraper.TweetID(val)

	data := NewTweetDetailData()
	data.MainTweetID = tweet_id

	is_scrape_required := r.URL.Query().Has("scrape")
	is_conversation_required := len(parts) <= 2 || (parts[2] != "like" && parts[2] != "unlike")

	tweet, err := app.ensure_tweet(tweet_id, is_scrape_required, is_conversation_required)
	var toasts []Toast
	if err != nil {
		app.ErrorLog.Print(fmt.Errorf("TweetDetail (%d): %w", tweet_id, err))
		if errors.Is(err, ErrNotFound) {
			// Can't find the tweet; abort
			app.error_404(w, r)
			return
		} else if errors.Is(err, scraper.ErrSessionInvalidated) {
			toasts = append(toasts, Toast{
				Title:   "Session invalidated",
				Message: "Your session has been invalidated by Twitter.  You'll have to log in again.",
				Type:    "error",
			})
			// TODO: delete the invalidated session
		} else if errors.Is(err, scraper.ErrRateLimited) {
			toasts = append(toasts, Toast{
				Title:   "Rate limited",
				Message: "While scraping, a rate-limit was hit.  Results may be incomplete.",
				Type:    "warning",
			})
		} else {
			panic(err) // Let the 500 handler deal with it
		}
	}

	req_with_tweet := r.WithContext(add_tweet_to_context(r.Context(), tweet))

	if len(parts) > 2 && parts[2] == "like" {
		app.LikeTweet(w, req_with_tweet)
		return
	} else if len(parts) > 2 && parts[2] == "unlike" {
		app.UnlikeTweet(w, req_with_tweet)
		return
	}

	twt_detail, err := app.Profile.GetTweetDetail(data.MainTweetID, app.ActiveUser.ID)
	panic_if(err) // ErrNotInDatabase should be impossible, since we already fetched the single tweet successfully

	data.TweetDetailView = twt_detail

	app.buffered_render_page(
		w,
		"tpl/tweet_detail.tpl",
		PageGlobalData{TweetTrove: twt_detail.TweetTrove, FocusedTweetID: data.MainTweetID, Toasts: toasts},
		data,
	)
}

type key string

const TWEET_KEY = key("tweet")

func add_tweet_to_context(ctx context.Context, tweet scraper.Tweet) context.Context {
	return context.WithValue(ctx, TWEET_KEY, tweet)
}

func get_tweet_from_context(ctx context.Context) scraper.Tweet {
	tweet, is_ok := ctx.Value(TWEET_KEY).(scraper.Tweet)
	if !is_ok {
		panic("Tweet not found in context")
	}
	return tweet
}
