package persistence_test

import (
	"time"
	"fmt"
	"math/rand"

	"offline_twitter/scraper"
	"offline_twitter/persistence"
)

/**
 * Load a test profile, or create it if it doesn't exist
 */
func create_or_load_profile(profile_path string) persistence.Profile {
	var profile persistence.Profile
	var err error

	if !file_exists(profile_path) {
		profile, err = persistence.NewProfile(profile_path)
	} else {
		profile, err = persistence.LoadProfile(profile_path)
	}
	if err != nil {
		panic(err)
	}
	return profile
}

/**
 * Create a new user with a random ID and handle
 */
func create_dummy_user() scraper.User {
	rand.Seed(time.Now().UnixNano())
	userID := fmt.Sprint(rand.Int())

	return scraper.User{
		ID: scraper.UserID(userID),
		DisplayName: "display name",
		Handle: scraper.UserHandle("handle" + userID),
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
		PinnedTweetID: scraper.TweetID("234"),
	}
}


/**
 * Create a new tweet with a random ID and content
 */
func create_dummy_tweet() scraper.Tweet {
	rand.Seed(time.Now().UnixNano())
	tweet_id := fmt.Sprint(rand.Int())

	return scraper.Tweet{
		ID: scraper.TweetID(tweet_id),
		UserID: "user",
		Text: "text",
		PostedAt: time.Now().Truncate(1e9),  // Round to nearest second
		NumLikes: 1,
		NumRetweets: 2,
		NumReplies: 3,
		NumQuoteTweets: 4,
		Video: "video",
		Urls: []string{"url1", "url2"},
		Images: []string{"image1", "image2"},
		Mentions: []scraper.UserHandle{"mention1", "mention2"},
		Hashtags: []string{"hash1", "hash2"},
	}
}
