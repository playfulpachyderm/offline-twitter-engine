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
		if err := recover(); err != nil {
			// TODO
			fmt.Println("Panicked!")
			fmt.Printf("%#v\n", err)
		}
	}()

	fmt.Println("Starting scrape...")

	// Do nothing if scraping is currently disabled
	if app.IsScrapingDisabled {
		fmt.Println("Skipping scrape!")
		return
	}

	fmt.Println("Scraping...")
	trove, err := scraper.GetHomeTimeline("", is_for_you_only)
	if err != nil {
		app.ErrorLog.Printf("Background scrape failed: %s", err.Error())
		return
	}
	fmt.Println("Saving scrape results...")
	app.Profile.SaveTweetTrove(trove)
	fmt.Println("Scraping succeeded.")
	is_for_you_only = false
}

func (app *Application) start_background() {
	// Start a goroutine to run the background task every 3 minutes
	fmt.Println("Starting background")
	go func() {
		fmt.Println("Starting routine")

		// Initial delay before the first task execution (0 seconds here, adjust as needed)
		initialDelay := 10 * time.Second
		time.Sleep(initialDelay)

		app.background_scrape()

		// Create a timer that triggers the background task every 3 minutes
		interval := 3 * time.Minute // TODO: parameterizable
		timer := time.NewTicker(interval)
		defer timer.Stop()

		for range timer.C {
			// Execute the background task
			fmt.Println("Starting routine")

			app.background_scrape()
		}
	}()
}
