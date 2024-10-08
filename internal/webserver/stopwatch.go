package webserver

import (
	"fmt"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
	"time"
)

var is_following_only = true // Do one initial scrape of the "following_only" feed and then just regular feed after that

func (app *Application) background_scrape() {
	// Avoid crashing the thread if a scrape fails
	defer func() {
		if r := recover(); r != nil {
			// TODO
			fmt.Println("Background Home Timeline thread: panicked!")
			if err, ok := r.(error); ok {
				fmt.Println(err.Error())
			} else {
				fmt.Println(r)
			}
		}
	}()

	fmt.Println("Starting home timeline scrape...")

	// Do nothing if scraping is currently disabled
	if app.IsScrapingDisabled {
		fmt.Println("Skipping home timeline scrape!")
		return
	}

	fmt.Println("Scraping home timeline...")
	trove, err := app.API.GetHomeTimeline("", is_following_only)
	if err != nil {
		app.ErrorLog.Printf("Background scrape failed: %s", err.Error())
		return
	}
	fmt.Println("Saving scrape results...")
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)
	fmt.Println("Scraping succeeded.")
	is_following_only = false
}

func (app *Application) background_user_likes_scrape() {
	// Avoid crashing the thread if a scrape fails
	defer func() {
		if r := recover(); r != nil {
			// TODO
			fmt.Println("Background Home Timeline thread: panicked!")
			if err, ok := r.(error); ok {
				fmt.Println(err.Error())
			} else {
				fmt.Println(r)
			}
		}
	}()

	fmt.Println("Starting user likes scrape...")

	// Do nothing if scraping is currently disabled
	if app.IsScrapingDisabled {
		fmt.Println("Skipping user likes scrape!")
		return
	}

	fmt.Println("Scraping user likes...")
	trove, err := app.API.GetUserLikes(app.ActiveUser.ID, 50) // TODO: parameterizable
	if err != nil {
		app.ErrorLog.Printf("Background scrape failed: %s", err.Error())
		return
	}
	fmt.Println("Saving scrape results...")
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)
	fmt.Println("Scraping succeeded.")
}

var inbox_cursor string = ""

func (app *Application) background_dm_polling_scrape() {
	// Avoid crashing the thread if a scrape fails
	defer func() {
		if r := recover(); r != nil {
			// TODO
			fmt.Println("Background Home Timeline thread: panicked!")
			if err, ok := r.(error); ok {
				fmt.Println(err.Error())
			} else {
				fmt.Println(r)
			}
		}
	}()

	fmt.Println("Starting user DMs scrape...")

	// Do nothing if scraping is currently disabled
	if app.IsScrapingDisabled {
		fmt.Println("Skipping user DMs scrape!")
		return
	}

	fmt.Println("Scraping user DMs...")
	var trove scraper.TweetTrove
	var err error
	if inbox_cursor == "" {
		trove, inbox_cursor, err = app.API.GetInbox(0)
	} else {
		trove, inbox_cursor, err = app.API.PollInboxUpdates(inbox_cursor)
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("Saving DM results...")
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)
	fmt.Println("Scraping DMs succeeded.")
}

func (app *Application) background_notifications_scrape() {
	// Avoid crashing the thread if a scrape fails
	defer func() {
		if r := recover(); r != nil {
			// TODO
			fmt.Println("Background notifications thread: panicked!")
			if err, ok := r.(error); ok {
				fmt.Println(err.Error())
			} else {
				fmt.Println(r)
			}
		}
	}()

	fmt.Println("Starting notifications scrape...")

	// Do nothing if scraping is currently disabled
	if app.IsScrapingDisabled {
		fmt.Println("Skipping notifications scrape!")
		return
	}

	fmt.Println("Scraping user notifications...")
	trove, last_unread_notification_sort_index, err := app.API.GetNotifications(1) // Just 1 page
	if err != nil {
		panic(err)
	}
	// Jot down the unread notifs info in the application object (to render notification count bubble)
	app.LastReadNotificationSortIndex = last_unread_notification_sort_index
	fmt.Println("Saving notification results...")
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)
	fmt.Println("Scraping notification succeeded.")
}

func (app *Application) start_background() {
	fmt.Println("Starting background")

	// Scrape the home timeline every 3 minutes
	go func() {
		// Initial delay before the first task execution
		time.Sleep(10 * time.Second)
		app.background_scrape()

		// Create a timer that triggers the background task every 3 minutes
		interval := 3 * time.Minute // TODO: parameterizable
		timer := time.NewTicker(interval)
		defer timer.Stop()

		for range timer.C {
			app.background_scrape()
		}
	}()

	// Scrape the logged-in user's likes every 10 minutes
	go func() {
		time.Sleep(15 * time.Second)
		app.background_user_likes_scrape()

		interval := 10 * time.Minute // TODO: parameterizable
		timer := time.NewTicker(interval)
		defer timer.Stop()

		for range timer.C {
			app.background_user_likes_scrape()
		}
	}()

	// Scrape inbox DMs every 10 seconds
	go func() {
		time.Sleep(5 * time.Second)
		app.background_dm_polling_scrape()

		interval := 10 * time.Second
		timer := time.NewTicker(interval)
		defer timer.Stop()
		for range timer.C {
			app.background_dm_polling_scrape()
		}
	}()

	// Scrape notifications every 10 seconds
	go func() {
		app.background_notifications_scrape()

		interval := 10 * time.Second
		timer := time.NewTicker(interval)
		defer timer.Stop()
		for range timer.C {
			app.background_notifications_scrape()
		}
	}()
}
