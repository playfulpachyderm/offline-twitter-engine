package webserver

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"

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
	_, tail := path.Split(r.URL.Path)
	val, err := strconv.Atoi(tail)
	if err != nil {
		app.error_400_with_message(w, fmt.Sprintf("Invalid tweet ID: %q", tail))
		return
	}
	tweet_id := scraper.TweetID(val)

	data := NewTweetDetailData()
	data.MainTweetID = tweet_id

	// Return whether the scrape succeeded (if false, we should 404)
	try_scrape_tweet := func() bool {
		if app.DisableScraping {
			return false
		}
		trove, err := scraper.GetTweetFullAPIV2(tweet_id, 50) // TODO: parameterizable
		if err != nil {
			app.ErrorLog.Print(err)
			return false
		}
		app.Profile.SaveTweetTrove(trove)
		return true
	}

	tweet, err := app.Profile.GetTweetById(tweet_id)
	if err != nil {
		if errors.Is(err, persistence.ErrNotInDB) {
			if !try_scrape_tweet() {
				app.error_404(w)
				return
			}
		} else {
			panic(err)
		}
	} else if !tweet.IsConversationScraped {
		try_scrape_tweet() // If it fails, we can still render it (not 404)
	}

	trove, err := app.Profile.GetTweetDetail(data.MainTweetID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotInDB) {
			app.error_404(w)
			return
		} else {
			panic(err)
		}
	}
	data.TweetDetailView = trove

	app.buffered_render_tweet_page(w, "tpl/tweet_detail.tpl", data)
}
