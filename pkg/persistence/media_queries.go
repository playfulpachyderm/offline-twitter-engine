package persistence

import (
	"fmt"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Save an Image
//
// args:
// - img: the Image to save
func (p Profile) SaveImage(img Image) error {
	_, err := p.DB.NamedExec(`
		insert into images (id, tweet_id, width, height, remote_url, local_filename, is_downloaded)
		            values (:id, :tweet_id, :width, :height, :remote_url, :local_filename, :is_downloaded)
		       on conflict do update
		               set is_downloaded=(is_downloaded or :is_downloaded)
		`,
		img,
	)
	if err != nil {
		return fmt.Errorf("Error saving image (tweet ID %d):\n  %w", img.TweetID, err)
	}
	return nil
}

// Save a Video
//
// args:
// - img: the Video to save
func (p Profile) SaveVideo(vid Video) error {
	_, err := p.DB.NamedExec(`
		insert into videos (id, tweet_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename,
		                    duration, view_count, is_downloaded, is_blocked_by_dmca, is_gif)
		            values (:id, :tweet_id, :width, :height, :remote_url, :local_filename, :thumbnail_remote_url, :thumbnail_local_filename,
                            :duration, :view_count, :is_downloaded, :is_blocked_by_dmca, :is_gif)
		       on conflict do update
		               set is_downloaded=(is_downloaded or :is_downloaded),
		                   view_count=max(view_count, :view_count),
						   is_blocked_by_dmca = :is_blocked_by_dmca
		`,
		vid,
	)
	if err != nil {
		return fmt.Errorf("Error saving video (tweet ID %d):\n  %w", vid.TweetID, err)
	}
	return nil
}

// Save an Url
func (p Profile) SaveUrl(url Url) error {
	_, err := p.DB.NamedExec(`
		insert into urls (tweet_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
		                  thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded)
		          values (:tweet_id, :domain, :text, :short_text, :title, :description, :creator_id, :site_id, :thumbnail_width,
                          :thumbnail_height, :thumbnail_remote_url, :thumbnail_local_path, :has_card, :has_thumbnail, :is_content_downloaded
                         )
		     on conflict do update
		             set is_content_downloaded=(is_content_downloaded or :is_content_downloaded)
		`,
		url,
	)
	if err != nil {
		return fmt.Errorf("Error saving Url (tweet ID %d):\n  %w", url.TweetID, err)
	}
	return nil
}

// Save a Poll
func (p Profile) SavePoll(poll Poll) error {
	_, err := p.DB.NamedExec(`
		insert into polls (id, tweet_id, num_choices, choice1, choice1_votes, choice2, choice2_votes, choice3, choice3_votes, choice4,
		                   choice4_votes, voting_duration, voting_ends_at, last_scraped_at)
		           values (:id, :tweet_id, :num_choices, :choice1, :choice1_votes, :choice2, :choice2_votes, :choice3, :choice3_votes,
                           :choice4, :choice4_votes, :voting_duration, :voting_ends_at, :last_scraped_at)
		      on conflict do update
		              set choice1_votes=:choice1_votes,
		                  choice2_votes=:choice2_votes,
		                  choice3_votes=:choice3_votes,
		                  choice4_votes=:choice4_votes,
		                  last_scraped_at=:last_scraped_at
		`,
		poll,
	)
	if err != nil {
		return fmt.Errorf("Error saving Poll (tweet ID %d):\n  %w", poll.TweetID, err)
	}
	return nil
}

// Get the list of images for a tweet
func (p Profile) GetImagesForTweet(t Tweet) (imgs []Image, err error) {
	err = p.DB.Select(&imgs,
		"select id, tweet_id, width, height, remote_url, local_filename, is_downloaded from images where tweet_id=?",
		t.ID)
	return
}

// Get the list of videos for a tweet
func (p Profile) GetVideosForTweet(t Tweet) (vids []Video, err error) {
	err = p.DB.Select(&vids, `
		select id, tweet_id, width, height, remote_url, local_filename, thumbnail_remote_url, thumbnail_local_filename, duration,
		       view_count, is_downloaded, is_blocked_by_dmca, is_gif
		  from videos
		 where tweet_id = ?
	`, t.ID)
	return
}

// Get the list of Urls for a Tweet
func (p Profile) GetUrlsForTweet(t Tweet) (urls []Url, err error) {
	err = p.DB.Select(&urls, `
		select tweet_id, domain, text, short_text, title, description, creator_id, site_id, thumbnail_width, thumbnail_height,
		       thumbnail_remote_url, thumbnail_local_path, has_card, has_thumbnail, is_content_downloaded
		  from urls
		 where tweet_id = ?
		 order by rowid
	`, t.ID)
	return
}

// Get the list of Polls for a Tweet
func (p Profile) GetPollsForTweet(t Tweet) (polls []Poll, err error) {
	err = p.DB.Select(&polls, `
		select id, tweet_id, num_choices, choice1, choice1_votes, choice2, choice2_votes, choice3, choice3_votes, choice4, choice4_votes,
		       voting_duration, voting_ends_at, last_scraped_at
		  from polls
		 where tweet_id = ?
	`, t.ID)
	return
}
