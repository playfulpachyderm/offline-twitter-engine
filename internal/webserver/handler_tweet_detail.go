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
		if errors.Is(err, persistence.ErrNotInDB) {
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
		trove, err := scraper.GetTweetFullAPIV2(id, 50) // TODO: parameterizable
		if err == nil || errors.Is(err, scraper.END_OF_FEED) || errors.Is(err, scraper.ErrRateLimited) {
			app.Profile.SaveTweetTrove(trove, false)
			go app.Profile.SaveTweetTrove(trove, true) // Download the content in the background
			_, is_available = trove.Tweets[id]
		} else {
			app.ErrorLog.Print(err)
			// TODO: show error in UI
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
	like, err := scraper.LikeTweet(tweet.ID)
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
	err := scraper.UnlikeTweet(tweet.ID)
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
		app.error_400_with_message(w, fmt.Sprintf("Invalid tweet ID: %q", parts[1]))
		return
	}
	tweet_id := scraper.TweetID(val)

	data := NewTweetDetailData()
	data.MainTweetID = tweet_id

	is_scrape_required := r.URL.Query().Has("scrape")
	is_conversation_required := len(parts) <= 2 || (parts[2] != "like" && parts[2] != "unlike")

	tweet, err := app.ensure_tweet(tweet_id, is_scrape_required, is_conversation_required)
	if errors.Is(err, ErrNotFound) {
		app.error_404(w)
		return
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
	panic_if(err) // ErrNotInDB should be impossible, since we already fetched the single tweet successfully

	data.TweetDetailView = twt_detail

	app.buffered_render_page(
		w,
		"tpl/tweet_detail.tpl",
		PageGlobalData{TweetTrove: twt_detail.TweetTrove, FocusedTweetID: data.MainTweetID},
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
