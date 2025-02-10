package webserver

import (
	"errors"
	"fmt"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// DUPE: full_save_tweet_trove
func (app *Application) full_save_tweet_trove(trove TweetTrove) {
	// Save the initial trove
	conflicting_users := app.Profile.SaveTweetTrove(trove, false, app.API.DownloadMedia)

	// Handle conflicting users
	for _, u_id := range conflicting_users {
		app.InfoLog.Printf("Conflicting user handle found (ID %d); old user has been marked deleted.  Rescraping manually", u_id)
		// Rescrape
		updated_user, err := GetUserByID(u_id)
		if errors.Is(err, ErrDoesntExist) {
			// Mark them as deleted.
			// Handle and display name won't be updated if the user exists.
			updated_user = User{ID: u_id, DisplayName: "<Unknown User>", Handle: "<UNKNOWN USER>", IsDeleted: true}
		} else if err != nil {
			panic(fmt.Errorf("error scraping conflicting user (ID %d): %w", u_id, err))
		}
		err = app.Profile.SaveUser(&updated_user)
		if err != nil {
			panic(fmt.Errorf(
				"error saving rescraped conflicting user with ID %d and handle %q: %w",
				updated_user.ID, updated_user.Handle, err,
			))
		}
	}

	// Download media content in background
	go app.Profile.SaveTweetTrove(trove, true, app.API.DownloadMedia)
}
