package webserver

import (
	"fmt"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
	"time"
)

var is_for_you_only = true // Do one initial scrape of the "for you" feed and then just regular feed after that

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
	trove, err := scraper.GetHomeTimeline("", is_for_you_only)
	if err != nil {
		app.ErrorLog.Printf("Background scrape failed: %s", err.Error())
		return
	}
	fmt.Println("Saving scrape results...")
	app.Profile.SaveTweetTrove(trove, false)
	go app.Profile.SaveTweetTrove(trove, true)
	fmt.Println("Scraping succeeded.")
	is_for_you_only = false
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
	trove, err := scraper.GetUserLikes(app.ActiveUser.ID, 50) // TODO: parameterizable
	if err != nil {
		app.ErrorLog.Printf("Background scrape failed: %s", err.Error())
		return
	}
	fmt.Println("Saving scrape results...")
	app.Profile.SaveTweetTrove(trove, false)
	go app.Profile.SaveTweetTrove(trove, true)
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
	var trove scraper.DMTrove
	if inbox_cursor == "" {
		trove, inbox_cursor = scraper.GetInbox(0)
	} else {
		trove, inbox_cursor = scraper.PollInboxUpdates(inbox_cursor)
	}
	fmt.Println("Saving DM results...")
	app.Profile.SaveDMTrove(trove, false)
	go app.Profile.SaveDMTrove(trove, true)
	fmt.Println("Scraping DMs succeeded.")
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
}
