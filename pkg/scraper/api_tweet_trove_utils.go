package scraper

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func (api *API) FillSpaceDetails(trove *TweetTrove) error {
	fmt.Println("Filling space details")
	for i := range trove.Spaces {
		fmt.Printf("Getting space: %q\n", trove.Spaces[i].ID)
		new_trove, err := api.FetchSpaceDetail(trove.Spaces[i].ID)
		if err != nil {
			return err
		}
		// Replace the old space in the trove with the new, updated one
		new_space, is_ok := new_trove.Spaces[i]
		if new_space.ShortUrl == "" {
			// Copy over the short-url, which doesn't seem to exist on a full Space response
			new_space.ShortUrl = trove.Spaces[i].ShortUrl
		}
		if is_ok {
			// Necessary to check is_ok because the space response could be empty, in which case
			// we don't want to overwrite it
			trove.Spaces[i] = new_space
		}
	}
	return nil
}

func (api *API) PostProcess(trove *TweetTrove) error {
	api.FetchTombstoneUsers(trove)
	trove.FillMissingUserIDs()
	err := api.FillSpaceDetails(trove)
	if err != nil {
		return err
	}
	return nil
}

// Tries to fetch every User that's been identified in a tombstone in this trove
func (api *API) FetchTombstoneUsers(trove *TweetTrove) {
	for _, handle := range trove.TombstoneUsers {
		// Skip fetching if this user is already in the trove
		user, already_fetched := trove.FindUserByHandle(handle)

		if already_fetched {
			// If the user is already fetched and it's an intact user, don't fetch it again
			if user.JoinDate.Unix() != (Timestamp{}).Unix() && user.JoinDate.Unix() != 0 {
				log.Debugf("Skipping %q due to intact user", handle)
				continue
			}

			// A user needs a valid handle or ID to fetch it by
			if user.IsIdFake && user.Handle == "<UNKNOWN USER>" {
				log.Debugf("Skipping %q due to completely unknown user (not fetchable)", handle)
				continue
			}
		}

		log.Debug("Getting tombstone user: " + handle)
		user, err := api.GetUser(handle)
		if errors.Is(err, ErrDoesntExist) {
			user = GetUnknownUserWithHandle(handle)
			user.IsDeleted = true
		} else if err != nil {
			panic(fmt.Errorf("Error getting tombstoned user with handle %q: \n  %w", handle, err))
		}

		if user.ID == 0 {
			// Find some random ID to fit it into the trove
			for i := 1; ; i++ {
				_, ok := trove.Users[UserID(i)]
				if !ok {
					user.ID = UserID(i)
					break
				}
			}
		}

		trove.Users[user.ID] = user
	}
}
