package persistence

import (
	"errors"
	"fmt"
	"path"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Convenience function that saves all the objects in a TweetTrove.
// Panics if anything goes wrong.
//
// TODO: a lot of this function contains duplicated code and should be extracted to functions
func (p Profile) SaveDMTrove(trove DMTrove, should_download bool) {
	p.SaveTweetTrove(trove.TweetTrove, should_download)

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
			downloader := DefaultDownloader{}

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
				if err != nil {
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

				if errors.Is(err, ErrorDMCA) {
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
				if err != nil {
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
					if err != nil {
						panic(fmt.Errorf("downloading link thumbnail %q on DM message %d:\n  %w", url.ThumbnailRemoteUrl, m.ID, err))
					}
				}
				url.IsContentDownloaded = true

				// Update it in the DB
				_, err = p.DB.NamedExec(`update chat_message_urls set is_content_downloaded = :is_content_downloaded where id = :id`, url)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
