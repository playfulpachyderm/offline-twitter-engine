package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type TweetDetailData struct {
	persistence.TweetDetailView
	MainTweetID scraper.TweetID
}

func NewTweetDetailData() TweetDetailData {
	return TweetDetailData{
		TweetDetailView: persistence.NewTweetDetailView(),
	}
}
func (t TweetDetailData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t TweetDetailData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}
func (t TweetDetailData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}
func (t TweetDetailData) Space(id scraper.SpaceID) scraper.Space {
	return t.Spaces[id]
}
func (t TweetDetailData) FocusedTweetID() scraper.TweetID {
	return t.MainTweetID
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

	is_needing_scrape := (len(parts) > 2 && parts[2] == "scrape")
	is_available := false

	// Check if tweet is already in DB
	tweet, err := app.Profile.GetTweetById(tweet_id)
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
	if is_available && len(parts) > 2 && (parts[2] == "like" || parts[2] == "unlike") {
		is_needing_scrape = false
	}

	if is_needing_scrape && !app.IsScrapingDisabled {
		trove, err := scraper.GetTweetFullAPIV2(tweet_id, 50) // TODO: parameterizable
		if err == nil {
			app.Profile.SaveTweetTrove(trove)
			is_available = true
		} else {
			app.ErrorLog.Print(err)
			// TODO: show error in UI
		}
	}

	if !is_available {
		app.error_404(w)
		return
	}

	if len(parts) > 2 && parts[2] == "like" {
		like, err := scraper.LikeTweet(tweet.ID)
		// if err != nil && !errors.Is(err, scraper.AlreadyLikedThisTweet) {}
		panic_if(err)
		fmt.Printf("Like: %#v\n", like)
		err = app.Profile.SaveLike(like)
		panic_if(err)
		tweet.IsLikedByCurrentUser = true

		app.buffered_render_basic_htmx(w, "likes-count", tweet)
		return
	} else if len(parts) > 2 && parts[2] == "unlike" {
		err = scraper.UnlikeTweet(tweet_id)
		panic_if(err)
		err = app.Profile.DeleteLike(scraper.Like{UserID: app.ActiveUser.ID, TweetID: tweet.ID})
		panic_if(err)
		tweet.IsLikedByCurrentUser = false

		app.buffered_render_basic_htmx(w, "likes-count", tweet)
		return
	}

	trove, err := app.Profile.GetTweetDetail(data.MainTweetID, app.ActiveUser.ID)
	panic_if(err) // ErrNotInDB should be impossible, since we already fetched the single tweet successfully

	data.TweetDetailView = trove
	// fmt.Println(to_json(data))

	app.buffered_render_tweet_page(w, "tpl/tweet_detail.tpl", data)
}
