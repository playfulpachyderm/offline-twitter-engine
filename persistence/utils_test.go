package persistence_test

import (
	"fmt"
	"math/rand"
	"time"

	"offline_twitter/persistence"
	"offline_twitter/scraper"
)

/**
 * Load a test profile, or create it if it doesn't exist.
 */
func create_or_load_profile(profile_path string) persistence.Profile {
	var profile persistence.Profile
	var err error

	if !file_exists(profile_path) {
		profile, err = persistence.NewProfile(profile_path)
		if err != nil {
			panic(err)
		}
		u := create_stable_user()
		err = profile.SaveUser(&u)
		if err != nil {
			panic(err)
		}
		err = profile.SaveTweet(create_stable_tweet())
		if err != nil {
			panic(err)
		}
		err = profile.SaveRetweet(create_stable_retweet())
	} else {
		profile, err = persistence.LoadProfile(profile_path)
	}
	if err != nil {
		panic(err)
	}
	return profile
}

/**
 * Create a stable user with a fixed ID and handle
 */
func create_stable_user() scraper.User {
	return scraper.User{
		ID:                    scraper.UserID(-1),
		DisplayName:           "stable display name",
		Handle:                scraper.UserHandle("handle stable"),
		Bio:                   "stable bio",
		FollowersCount:        10,
		FollowingCount:        2000,
		Location:              "stable location",
		Website:               "stable website",
		JoinDate:              scraper.TimestampFromUnix(10000000),
		IsVerified:            true,
		IsPrivate:             false,
		ProfileImageUrl:       "stable profile image url",
		ProfileImageLocalPath: "stable profile image local path",
		BannerImageUrl:        "stable banner image url",
		BannerImageLocalPath:  "stable image local path",
		PinnedTweetID:         scraper.TweetID(345),
	}
}

/**
 * Create a semi-stable Image based on the given ID
 */
func create_image_from_id(id int) scraper.Image {
	filename := fmt.Sprintf("image%d.jpg", id)
	return scraper.Image{
		ID:            scraper.ImageID(id),
		TweetID:       -1,
		Width:         id * 10,
		Height:        id * 5,
		RemoteURL:     filename,
		LocalFilename: filename,
		IsDownloaded:  false,
	}
}

/**
 * Create a semi-stable Video based on the given ID
 */
func create_video_from_id(id int) scraper.Video {
	filename := fmt.Sprintf("video%d.jpg", id)
	return scraper.Video{
		ID:                 scraper.VideoID(id),
		TweetID:            -1,
		Width:              id * 10,
		Height:             id * 5,
		RemoteURL:          filename,
		LocalFilename:      filename,
		ThumbnailRemoteUrl: filename,
		ThumbnailLocalPath: filename,
		Duration:           10000,
		ViewCount:          200,
		IsDownloaded:       false,
		IsGif:              false,
	}
}

/**
 * Create a semi-stable Url based on the given ID
 */
func create_url_from_id(id int) scraper.Url {
	s := fmt.Sprint(id)
	return scraper.Url{
		TweetID:             -1,
		Domain:              s + "domain",
		Text:                s + "text",
		ShortText:           s + "shorttext",
		Title:               s + "title",
		Description:         s + "description",
		ThumbnailWidth:      id * 23,
		ThumbnailHeight:     id * 7,
		ThumbnailRemoteUrl:  s + "remote url",
		ThumbnailLocalPath:  s + "local path",
		CreatorID:           scraper.UserID(id),
		SiteID:              scraper.UserID(id),
		HasCard:             true,
		IsContentDownloaded: false,
	}
}

/**
 * Create a semi-stable Poll based on the given ID
 */
func create_poll_from_id(id int) scraper.Poll {
	s := fmt.Sprint(id)
	return scraper.Poll{
		ID:             scraper.PollID(id),
		TweetID:        -1,
		NumChoices:     2,
		Choice1:        s,
		Choice1_Votes:  1000,
		Choice2:        "Not " + s,
		Choice2_Votes:  1500,
		VotingDuration: 10,
		VotingEndsAt:   scraper.TimestampFromUnix(10000000),
		LastUpdatedAt:  scraper.TimestampFromUnix(10000),
	}
}

/**
 * Create a stable tweet with a fixed ID and content
 */
func create_stable_tweet() scraper.Tweet {
	tweet_id := scraper.TweetID(-1)
	return scraper.Tweet{
		ID:             tweet_id,
		UserID:         -1,
		Text:           "stable text",
		PostedAt:       scraper.TimestampFromUnix(10000000),
		NumLikes:       10,
		NumRetweets:    10,
		NumReplies:     10,
		NumQuoteTweets: 10,
		Videos: []scraper.Video{
			create_video_from_id(-1),
		},
		Urls: []scraper.Url{
			create_url_from_id(-1),
		},
		Images: []scraper.Image{
			create_image_from_id(-1),
		},
		Mentions: scraper.CommaSeparatedList{},
		Hashtags: scraper.CommaSeparatedList{},
		Polls: []scraper.Poll{
			create_poll_from_id(-1),
		},
		Spaces: []scraper.Space{
			create_space_from_id(-1),
		},
		SpaceID:               scraper.SpaceID("some_id_-1"),
		IsConversationScraped: true,
		LastScrapedAt:         scraper.TimestampFromUnix(100000000),
	}
}

/**
 * Create a stable retweet with a fixed ID and parameters
 */
func create_stable_retweet() scraper.Retweet {
	retweet_id := scraper.TweetID(-1)
	return scraper.Retweet{
		RetweetID:     retweet_id,
		TweetID:       -1,
		RetweetedByID: -1,
		RetweetedAt:   scraper.TimestampFromUnix(20000000),
	}
}

/**
 * Create a new user with a random ID and handle
 */
func create_dummy_user() scraper.User {
	rand.Seed(time.Now().UnixNano())
	userID := rand.Int()

	return scraper.User{
		ID:                    scraper.UserID(userID),
		DisplayName:           "display name",
		Handle:                scraper.UserHandle(fmt.Sprintf("handle%d", userID)),
		Bio:                   "bio",
		FollowersCount:        0,
		FollowingCount:        1000,
		Location:              "location",
		Website:               "website",
		JoinDate:              scraper.Timestamp{time.Now().Truncate(1e9)}, // Round to nearest second
		IsVerified:            false,
		IsPrivate:             true,
		ProfileImageUrl:       "profile image url",
		ProfileImageLocalPath: "profile image local path",
		BannerImageUrl:        "banner image url",
		BannerImageLocalPath:  "banner image local path",
		PinnedTweetID:         scraper.TweetID(234),
	}
}

/**
 * Create a new tweet with a random ID and content
 */
func create_dummy_tweet() scraper.Tweet {
	rand.Seed(time.Now().UnixNano())
	tweet_id := scraper.TweetID(rand.Int())

	img1 := create_image_from_id(rand.Int())
	img1.TweetID = tweet_id
	img2 := create_image_from_id(rand.Int())
	img2.TweetID = tweet_id
	vid := create_video_from_id(rand.Int())
	vid.TweetID = tweet_id

	url1 := create_url_from_id(rand.Int())
	url1.TweetID = tweet_id
	url2 := create_url_from_id(rand.Int())
	url2.TweetID = tweet_id

	poll := create_poll_from_id(rand.Int())
	poll.TweetID = tweet_id

	space := create_space_from_id(rand.Int())
	space_id := space.ID

	return scraper.Tweet{
		ID:             tweet_id,
		UserID:         -1,
		Text:           "text",
		PostedAt:       scraper.Timestamp{time.Now().Truncate(1e9)}, // Round to nearest second
		NumLikes:       1,
		NumRetweets:    2,
		NumReplies:     3,
		NumQuoteTweets: 4,
		Videos:         []scraper.Video{vid},
		Urls:           []scraper.Url{url1, url2},
		Images:         []scraper.Image{img1, img2},
		Mentions:       scraper.CommaSeparatedList{"mention1", "mention2"},
		ReplyMentions:  scraper.CommaSeparatedList{"replymention1", "replymention2"},
		Hashtags:       scraper.CommaSeparatedList{"hash1", "hash2"},
		Polls:          []scraper.Poll{poll},
		Spaces:         []scraper.Space{space},
		SpaceID:        space_id,
	}
}

/**
 * Create a random tombstone
 */
func create_dummy_tombstone() scraper.Tweet {
	rand.Seed(time.Now().UnixNano())
	tweet_id := scraper.TweetID(rand.Int())

	return scraper.Tweet{
		ID:            tweet_id,
		UserID:        -1,
		TombstoneType: "deleted",
		IsStub:        true,
		Mentions:      scraper.CommaSeparatedList{},
		ReplyMentions: scraper.CommaSeparatedList{},
		Hashtags:      scraper.CommaSeparatedList{},
		Spaces:        []scraper.Space{},
	}
}

/**
 * Create a new retweet with a random ID for a given TweetID
 */
func create_dummy_retweet(tweet_id scraper.TweetID) scraper.Retweet {
	rand.Seed(time.Now().UnixNano())
	retweet_id := scraper.TweetID(rand.Int())

	return scraper.Retweet{
		RetweetID:     retweet_id,
		TweetID:       tweet_id,
		RetweetedByID: -1,
		RetweetedAt:   scraper.TimestampFromUnix(20000000),
	}
}

/**
 * Create a semi-stable Space given an ID
 */
func create_space_from_id(id int) scraper.Space {
	return scraper.Space{
		ID:             scraper.SpaceID(fmt.Sprintf("some_id_%d", id)),
		ShortUrl:       fmt.Sprintf("short_url_%d", id),
		State:          "Ended",
		Title:          "Some Title",
		CreatedAt:      scraper.TimestampFromUnix(1000),
		StartedAt:      scraper.TimestampFromUnix(2000),
		EndedAt:        scraper.TimestampFromUnix(3000),
		UpdatedAt:      scraper.TimestampFromUnix(4000),
		CreatedById:    -1,
		ParticipantIds: []scraper.UserID{-1},
	}
}
