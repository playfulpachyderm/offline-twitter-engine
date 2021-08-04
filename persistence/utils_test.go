package persistence_test

import (
	"time"
	"fmt"
	"math/rand"

	"offline_twitter/scraper"
	"offline_twitter/persistence"
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
		err = profile.SaveUser(create_stable_user())
		if err != nil {
			panic(err)
		}
		err = profile.SaveTweet(create_stable_tweet())
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
		ID: scraper.UserID(-1),
		DisplayName: "stable display name",
		Handle: scraper.UserHandle("handle stable"),
		Bio: "stable bio",
		FollowersCount: 10,
		FollowingCount: 2000,
		Location: "stable location",
		Website:"stable website",
		JoinDate: time.Unix(10000000, 0),
		IsVerified: true,
		IsPrivate: false,
		ProfileImageUrl: "stable profile image url",
		BannerImageUrl: "stable banner image url",
		PinnedTweetID: scraper.TweetID(345),
	}
}

/**
 * Create a semi-stable image based on the given ID
 */
func create_image_from_id(id int) scraper.Image {
	filename := fmt.Sprintf("image%d.jpg", id)
	return scraper.Image{
		ID: scraper.ImageID(id),
		TweetID: -1,
		Filename: filename,
		IsDownloaded: false,
	}
}

/**
 * Create a stable tweet with a fixed ID and content
 */
func create_stable_tweet() scraper.Tweet {
	tweet_id := scraper.TweetID(-1)
	return scraper.Tweet{
		ID: tweet_id,
		UserID: -1,
		Text: "stable text",
		PostedAt: time.Unix(10000000, 0),
		NumLikes: 10,
		NumRetweets: 10,
		NumReplies: 10,
		NumQuoteTweets: 10,
		Videos: []scraper.Video{{ID: scraper.VideoID(1), TweetID: tweet_id, Filename: "asdf", IsDownloaded: false}},
		Urls: []string{},
		Images: []scraper.Image{
			create_image_from_id(-1),
		},
		Mentions: []scraper.UserHandle{},
		Hashtags: []string{},
	}
}


/**
 * Create a new user with a random ID and handle
 */
func create_dummy_user() scraper.User {
	rand.Seed(time.Now().UnixNano())
	userID := rand.Int()

	return scraper.User{
		ID: scraper.UserID(userID),
		DisplayName: "display name",
		Handle: scraper.UserHandle(fmt.Sprintf("handle%d", userID)),
		Bio: "bio",
		FollowersCount: 0,
		FollowingCount: 1000,
		Location: "location",
		Website:"website",
		JoinDate: time.Now().Truncate(1e9),  // Round to nearest second
		IsVerified: false,
		IsPrivate: true,
		ProfileImageUrl: "profile image url",
		BannerImageUrl: "banner image url",
		PinnedTweetID: scraper.TweetID(234),
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

	return scraper.Tweet{
		ID: tweet_id,
		UserID: -1,
		Text: "text",
		PostedAt: time.Now().Truncate(1e9),  // Round to nearest second
		NumLikes: 1,
		NumRetweets: 2,
		NumReplies: 3,
		NumQuoteTweets: 4,
		Videos: []scraper.Video{scraper.Video{TweetID: tweet_id, Filename: "video" + fmt.Sprint(tweet_id), IsDownloaded: false}},
		Urls: []string{"url1", "url2"},
		Images: []scraper.Image{img1, img2},
		Mentions: []scraper.UserHandle{"mention1", "mention2"},
		Hashtags: []string{"hash1", "hash2"},
	}
}
