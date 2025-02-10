package webserver

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type BackgroundTask struct {
	Name         string
	GetTroveFunc func(*scraper.API) scraper.TweetTrove
	StartDelay   time.Duration
	Period       time.Duration

	log *log.Logger
	app *Application
}

func (t *BackgroundTask) Do() {
	// Avoid crashing the thread if a scrape fails
	defer func() {
		if r := recover(); r != nil {
			// TODO
			t.log.Print("panicked!")
			if err, ok := r.(error); ok {
				t.log.Print("(the following is an error)")
				t.log.Print(err.Error())
			} else {
				t.log.Print("(the following is an object, not an error)")
				t.log.Print(r)
			}
			panic_if(t.log.Output(2, string(debug.Stack())))
		}
	}()

	// Do nothing if scraping is currently disabled
	if t.app.IsScrapingDisabled {
		t.log.Print("(disabled)")
		return
	} else {
		t.log.Print("starting scrape")
	}

	// Run the task
	trove := t.GetTroveFunc(&t.app.API)
	t.log.Print("saving results")
	t.app.full_save_tweet_trove(trove)
	t.log.Print("success")
}

func (t *BackgroundTask) StartBackground() {
	// Start the task in a goroutine
	t.log = log.New(os.Stdout, fmt.Sprintf("[background (%s)]: ", t.Name), log.LstdFlags)

	go func() {
		t.log.Printf("starting, with initial delay %s and regular delay %s", t.StartDelay, t.Period)

		time.Sleep(t.StartDelay)          // Initial delay
		timer := time.NewTicker(t.Period) // Regular delay
		defer timer.Stop()

		t.Do()
		for range timer.C {
			t.Do()
		}
	}()
}

var is_following_only = 0           // Do mostly "For you" feed, but start with one round of the "following_only" feed
var is_following_only_frequency = 5 // Make every 5th scrape a "following_only" one

var inbox_cursor string = ""

func (app *Application) start_background() {
	fmt.Println("Starting background tasks")

	timeline_task := BackgroundTask{
		Name: "home timeline",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			should_do_following_only := is_following_only%is_following_only_frequency == 0
			trove, err := api.GetHomeTimeline("", should_do_following_only)
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			return trove
		},
		StartDelay: 10 * time.Second,
		Period:     3 * time.Minute,
		app:        app,
	}
	timeline_task.StartBackground()

	likes_task := BackgroundTask{
		Name: "user likes",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			trove, err := api.GetUserLikes(api.UserID, 50) // TODO: parameterizable
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			return trove
		},
		StartDelay: 15 * time.Second,
		Period:     10 * time.Minute,
		app:        app,
	}
	likes_task.StartBackground()

	dms_task := BackgroundTask{
		Name: "DM inbox",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			var trove scraper.TweetTrove
			var err error
			if inbox_cursor == "" {
				trove, inbox_cursor, err = api.GetInbox(0)
			} else {
				trove, inbox_cursor, err = api.PollInboxUpdates(inbox_cursor)
			}
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			return trove
		},
		StartDelay: 5 * time.Second,
		Period:     10 * time.Second,
		app:        app,
	}
	dms_task.StartBackground()

	notifications_task := BackgroundTask{
		Name: "DM inbox",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			trove, last_unread_notification_sort_index, err := api.GetNotifications(1) // Just 1 page
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			// Jot down the unread notifs info in the application object (to render notification count bubble)
			app.LastReadNotificationSortIndex = last_unread_notification_sort_index
			return trove
		},
		StartDelay: 1 * time.Second,
		Period:     10 * time.Second,
		app:        app,
	}
	notifications_task.StartBackground()

	bookmarks_task := BackgroundTask{
		Name: "bookmarks",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			trove, err := app.API.GetBookmarks(10)
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			return trove
		},
		StartDelay: 5 * time.Second,
		Period:     10 * time.Minute,
		app:        app,
	}
	bookmarks_task.StartBackground()

	own_profile_task := BackgroundTask{
		Name: "user profile",
		GetTroveFunc: func(api *scraper.API) scraper.TweetTrove {
			trove, err := app.API.GetUserFeed(api.UserID, 1)
			if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
				panic(err)
			}
			return trove
		},
		StartDelay: 1 * time.Second,
		Period:     20 * time.Minute,
		app:        app,
	}
	own_profile_task.StartBackground()
}
