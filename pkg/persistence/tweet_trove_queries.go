package persistence

import (
	"errors"
	"fmt"
	"path"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Convenience function that saves all the objects in a TweetTrove.
// Panics if anything goes wrong.
func (p Profile) SaveTweetTrove(trove TweetTrove, should_download bool, api *API) {
	for i, u := range trove.Users {
		err := p.SaveUser(&u)
		// Check for handle conflicts and handle them in place
		// TODO: this is hacky, it doesn't go here.  We should return a list of conflicting users
		// who were marked as deleted, and then let the callee re-scrape and re-save them.
		var conflict_err ErrConflictingUserHandle
		if errors.As(err, &conflict_err) {
			fmt.Printf(
				"Conflicting user handle found (ID %d); old user has been marked deleted.  Rescraping them\n",
				conflict_err.ConflictingUserID,
			)
			user, err := GetUserByID(conflict_err.ConflictingUserID)
			if errors.Is(err, ErrDoesntExist) {
				// Mark them as deleted.
				// Handle and display name won't be updated if the user exists.
				user = User{ID: conflict_err.ConflictingUserID, DisplayName: "<Unknown User>", Handle: "<UNKNOWN USER>", IsDeleted: true}
			} else if err != nil {
				panic(fmt.Errorf("error scraping conflicting user (ID %d): %w", conflict_err.ConflictingUserID, err))
			}
			err = p.SaveUser(&user)
			if err != nil {
				panic(fmt.Errorf("error saving rescraped conflicting user with ID %d and handle %q: %w", user.ID, user.Handle, err))
			}
		} else if err != nil {
			panic(fmt.Errorf("Error saving user with ID %d and handle %s:\n  %w", u.ID, u.Handle, err))
		}
		fmt.Println(u.Handle, u.ID)
		// If the User's ID was updated in saving (i.e., Unknown User), update it in the Trove too
		// Also update tweets, retweets and spaces that reference this UserID
		for j, tweet := range trove.Tweets {
			if tweet.UserID == trove.Users[i].ID {
				tweet.UserID = u.ID
				trove.Tweets[j] = tweet
			}
		}
		for j, retweet := range trove.Retweets {
			if retweet.RetweetedByID == trove.Users[i].ID {
				retweet.RetweetedByID = u.ID
				trove.Retweets[j] = retweet
			}
		}
		for j, space := range trove.Spaces {
			if space.CreatedById == trove.Users[i].ID {
				space.CreatedById = u.ID
				trove.Spaces[j] = space
			}
		}
		trove.Users[i] = u

		if should_download {
			// Download their tiny profile image
			err = p.DownloadUserProfileImageTiny(&u, api)
			if errors.Is(err, ErrRequestTimeout) {
				// Forget about it; if it's important someone will try again
				fmt.Printf("Failed to @%s's tiny profile image (%q): %s\n", u.Handle, u.ProfileImageUrl, err.Error())
			} else if err != nil {
				panic(fmt.Errorf("Error downloading user content for user with ID %d and handle %s:\n  %w", u.ID, u.Handle, err))
			}
		}
	}

	for _, s := range trove.Spaces {
		err := p.SaveSpace(s)
		if err != nil {
			panic(fmt.Errorf("Error saving space with ID %s:\n  %w", s.ID, err))
		}
	}

	for _, t := range trove.Tweets {
		err := p.SaveTweet(t)
		if err != nil {
			panic(fmt.Errorf("Error saving tweet ID %d:\n  %w", t.ID, err))
		}

		if should_download {
			err = p.DownloadTweetContentFor(&t, api)
			if errors.Is(err, ErrRequestTimeout) || errors.Is(err, ErrMediaDownload404) {
				// Forget about it; if it's important someone will try again
				fmt.Printf("Failed to download tweet ID %d: %s\n", t.ID, err.Error())
			} else if err != nil {
				panic(fmt.Errorf("Error downloading tweet content for tweet ID %d:\n  %w", t.ID, err))
			}
		}
	}

	for _, r := range trove.Retweets {
		err := p.SaveRetweet(r)
		if err != nil {
			panic(fmt.Errorf("Error saving retweet with ID %d from user ID %d:\n  %w", r.RetweetID, r.RetweetedByID, err))
		}
	}

	for _, l := range trove.Likes {
		err := p.SaveLike(l)
		if err != nil {
			panic(fmt.Errorf("Error saving Like: %#v\n  %w", l, err))
		}
	}

	for _, b := range trove.Bookmarks {
		err := p.SaveBookmark(b)
		if err != nil {
			panic(fmt.Errorf("Error saving Bookmark: %#v\n  %w", b, err))
		}
	}

	for _, n := range trove.Notifications {
		p.SaveNotification(n)
	}

	// DM related content
	// ------------------

	for _, r := range trove.Rooms {
		err := p.SaveChatRoom(r)
		if err != nil {
			panic(fmt.Errorf("Error saving chat room: %#v\n  %w", r, err))
		}
	}
	for _, m := range trove.Messages {
		err := p.SaveChatMessage(m)
		if err != nil {
			panic(fmt.Errorf("Error saving chat message: %#v\n  %w", m, err))
		}

		// TODO: all of this is very duplicated and should be refactored
		// Copied from media_download.go functions:
		// - download_tweet_image, download_tweet_video, download_link_thumbnail
		// - DownloadTweetContentWithInjector
		// Copied from tweet_queries.go functions:
		// - CheckTweetContentDownloadNeeded

		// Download content if needed
		if should_download {
			downloader := DefaultDownloader{API: api}

			for _, img := range m.Images {
				// Check if it's already downloaded
				var is_downloaded bool
				err := p.DB.Get(&is_downloaded, `select is_downloaded from chat_message_images where id = ?`, img.ID)
				if err != nil {
					panic(err)
				}
				if is_downloaded {
					// Already downloaded; skip
					continue
				}

				// DUPE: download-image
				outfile := path.Join(p.ProfileDir, "images", img.LocalFilename)
				err = downloader.Curl(img.RemoteURL, outfile)
				if errors.Is(err, ErrRequestTimeout) {
					// Forget about it; if it's important someone will try again
					fmt.Printf("Failed to download image %q: %s\n", img.RemoteURL, err.Error())
				} else if err != nil {
					panic(fmt.Errorf("downloading image %q on DM message %d:\n  %w", img.RemoteURL, m.ID, err))
				}
				_, err = p.DB.NamedExec(`update chat_message_images set is_downloaded = 1 where id = :id`, img)
				if err != nil {
					panic(err)
				}
			}

			for _, vid := range m.Videos {
				// Videos can be geoblocked, and the HTTP response isn't in JSON so it's hard to capture
				if vid.IsGeoblocked {
					continue
				}

				// Check if it's already downloaded
				var is_downloaded bool
				err := p.DB.Get(&is_downloaded, `select is_downloaded from chat_message_videos where id = ?`, vid.ID)
				if err != nil {
					panic(err)
				}
				if is_downloaded {
					// Already downloaded; skip
					continue
				}

				// DUPE: download-video
				// Download the video
				outfile := path.Join(p.ProfileDir, "videos", vid.LocalFilename)
				err = downloader.Curl(vid.RemoteURL, outfile)

				if errors.Is(err, ErrRequestTimeout) {
					// Forget about it; if it's important someone will try again
					fmt.Printf("Failed to download video %q: %s\n", vid.RemoteURL, err.Error())
				} else if errors.Is(err, ErrorDMCA) {
					vid.IsDownloaded = false
					vid.IsBlockedByDMCA = true
				} else if err != nil {
					panic(fmt.Errorf("downloading video %q on DM message %d:\n  %w", vid.RemoteURL, m.ID, err))
				} else {
					vid.IsDownloaded = true
				}

				// Download the thumbnail
				outfile = path.Join(p.ProfileDir, "video_thumbnails", vid.ThumbnailLocalPath)
				err = downloader.Curl(vid.ThumbnailRemoteUrl, outfile)
				if errors.Is(err, ErrRequestTimeout) {
					// Forget about it; if it's important someone will try again
					fmt.Printf("Failed to download video thumbnail %q: %s\n", vid.ThumbnailRemoteUrl, err.Error())
				} else if err != nil {
					panic(fmt.Errorf("Error downloading video thumbnail (DMMessageID %d):\n  %w", vid.DMMessageID, err))
				}

				// Update it in the DB
				_, err = p.DB.NamedExec(`
					update chat_message_videos set is_downloaded = :is_downloaded, is_blocked_by_dmca = :is_blocked_by_dmca where id = :id
				`, vid)
				if err != nil {
					panic(err)
				}
			}

			for _, url := range m.Urls {
				// DUPE: download-link-thumbnail
				if url.HasCard && url.HasThumbnail {
					outfile := path.Join(p.ProfileDir, "link_preview_images", url.ThumbnailLocalPath)
					err := downloader.Curl(url.ThumbnailRemoteUrl, outfile)
					if errors.Is(err, ErrRequestTimeout) {
						// Forget about it; if it's important someone will try again
						fmt.Printf("Failed to download link thumbnail %q: %s\n", url.ThumbnailRemoteUrl, err.Error())
					} else if err != nil {
						panic(fmt.Errorf("downloading link thumbnail %q on DM message %d:\n  %w", url.ThumbnailRemoteUrl, m.ID, err))
					}
				}
				url.IsContentDownloaded = true

				// Update it in the DB
				_, err = p.DB.NamedExec(`
					update chat_message_urls set is_content_downloaded = :is_content_downloaded where chat_message_id = :chat_message_id
				`, url)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
