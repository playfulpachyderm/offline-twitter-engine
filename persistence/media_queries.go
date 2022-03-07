package persistence

import (
	"offline_twitter/scraper"
)

/**
 * Save an Image
 *
 * args:
 * - img: the Image to save
 */
func (p Profile) SaveImage(img scraper.Image) error {
	_, err := p.DB.Exec(`
        insert into images (id, tweet_id, width, height, remote_url, local_filename, is_downloaded)
                    values (?, ?, ?, ?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=(is_downloaded or ?)
        `,
		img.ID, img.TweetID, img.Width, img.Height, img.RemoteURL, img.LocalFilename, img.IsDownloaded,
		img.IsDownloaded,
	)
	return err
}

/**
 * Save a Video
 *
 * args:
 * - img: the Video to save
 */
func (p Profile) SaveVideo(vid scraper.Video) error {
	_, err := p.DB.Exec(`
        insert into videos (id, tweet_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename,
                            duration, view_count, is_downloaded, is_gif)
                    values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
               on conflict do update
                       set is_downloaded=(is_downloaded or ?),
                           view_count=max(view_count, ?)
        `,
		vid.ID, vid.TweetID, vid.Width, vid.Height, vid.RemoteURL, vid.LocalFilename, vid.ThumbnailRemoteUrl, vid.ThumbnailLocalPath,
		vid.Duration, vid.ViewCount, vid.IsDownloaded, vid.IsGif,

		vid.IsDownloaded, vid.ViewCount,
	)
	return err
}

/**
 * Save an Url
 */
func (p Profile) SaveUrl(url scraper.Url) error {
	_, err := p.DB.Exec(`
        insert into urls (tweet_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
                          thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded)
                  values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
             on conflict do update
                     set is_content_downloaded=(is_content_downloaded or ?)
        `,
		url.TweetID, url.Domain, url.Text, url.ShortText, url.Title, url.Description, url.CreatorID, url.SiteID, url.ThumbnailWidth,
		url.ThumbnailHeight, url.ThumbnailRemoteUrl, url.ThumbnailLocalPath, url.HasCard, url.HasThumbnail, url.IsContentDownloaded,

		url.IsContentDownloaded,
	)
	return err
}

/**
 * Save a Poll
 */
func (p Profile) SavePoll(poll scraper.Poll) error {
	_, err := p.DB.Exec(`
        insert into polls (id, tweet_id, num_choices, choice1, choice1_votes, choice2, choice2_votes, choice3, choice3_votes, choice4,
                           choice4_votes, voting_duration, voting_ends_at, last_scraped_at)
                   values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
              on conflict do update
                      set choice1_votes=?,
                          choice2_votes=?,
                          choice3_votes=?,
                          choice4_votes=?,
                          last_scraped_at=?
        `,
		poll.ID, poll.TweetID, poll.NumChoices, poll.Choice1, poll.Choice1_Votes, poll.Choice2, poll.Choice2_Votes, poll.Choice3,
		poll.Choice3_Votes, poll.Choice4, poll.Choice4_Votes, poll.VotingDuration, poll.VotingEndsAt, poll.LastUpdatedAt,

		poll.Choice1_Votes, poll.Choice2_Votes, poll.Choice3_Votes, poll.Choice4_Votes, poll.LastUpdatedAt,
	)
	return err
}

/**
 * Get the list of images for a tweet
 */
func (p Profile) GetImagesForTweet(t scraper.Tweet) (imgs []scraper.Image, err error) {
	err = p.DB.Select(&imgs,
        "select id, tweet_id, width, height, remote_url, local_filename, is_downloaded from images where tweet_id=?",
    t.ID)
	return
}

/**
 * Get the list of videos for a tweet
 */
func (p Profile) GetVideosForTweet(t scraper.Tweet) (vids []scraper.Video, err error) {
	err = p.DB.Select(&vids, `
        select id, tweet_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename, duration,
               view_count, is_downloaded, is_gif
          from videos
         where tweet_id = ?
    `, t.ID)
    return
}

/**
 * Get the list of Urls for a Tweet
 */
func (p Profile) GetUrlsForTweet(t scraper.Tweet) (urls []scraper.Url, err error) {
	err = p.DB.Select(&urls, `
        select tweet_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
               thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded
          from urls
         where tweet_id = ?
         order by rowid
    `, t.ID)
    return
}

/**
 * Get the list of Polls for a Tweet
 */
func (p Profile) GetPollsForTweet(t scraper.Tweet) (polls []scraper.Poll, err error) {
	err = p.DB.Select(&polls, `
        select id, tweet_id, num_choices, choice1, choice1_votes, choice2, choice2_votes, choice3, choice3_votes, choice4, choice4_votes,
               voting_duration, voting_ends_at, last_scraped_at
          from polls
         where tweet_id = ?
    `, t.ID)
    return
}
