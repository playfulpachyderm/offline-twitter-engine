package main

import (
	"os"
	"strconv"
	"fmt"
	"flag"
	"offline_twitter/scraper"
	"offline_twitter/persistence"
)

/**
 * Global variable referencing the open data profile
 */
var profile persistence.Profile


/**
 * Main method
 */
func main() {
	if len(os.Args) < 3 {
		die("", true, 1)
	}

	profile_dir := flag.String("profile", ".", "TODO USAGE")
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		die("", true, 1)
	}

	operation := args[0]
	target := args[1]

	if operation == "create_profile" {
		create_profile(target)
		return
	}



	var err error
	profile, err = persistence.LoadProfile(*profile_dir)
	if err != nil {
		die("Could not load profile: " + err.Error(), true, 2)
	}

	switch (operation) {
	case "create_profile":
		create_profile(target)
	case "fetch_user":
		fetch_user(scraper.UserHandle(target))
	case "fetch_tweet_only":
		fetch_tweet_only(target)
	case "download_tweet_content":
		download_tweet_content(target)
	case "download_user_content":
		download_user_content(scraper.UserHandle(target))
	default:
		die("Invalid operation: " + operation, true, 3)
	}
}

/**
 * Create a data directory.
 *
 * args:
 * - target_dir: the location of the new data dir.
 */
func create_profile(target_dir string) {
	_, err := persistence.NewProfile(target_dir)
	if err != nil {
		panic(err)
	}
}

/**
 * Scrape a user and save it in the database.
 *
 * args:
 * - handle: e.g., "michaelmalice"
 */
func fetch_user(handle scraper.UserHandle) {
	if profile.UserExists(handle) {
		fmt.Println("User is already in database.  Updating user...")
	}
	user, err := scraper.GetUser(handle)
	if err != nil {
		die(err.Error(), false, -1)
	}
	fmt.Println(user)

	err = profile.SaveUser(user)
	if err != nil {
		die("Error saving user: " + err.Error(), false, 4)
	}
	fmt.Println("Saved the user.  Exiting successfully")
}

/**
 * Scrape a single tweet and save it in the database.
 *
 * args:
 * - tweet_url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 */
func fetch_tweet_only(tweet_url string) {
	tweet_id, err := extract_id_from(tweet_url)
	if err != nil {
		die(err.Error(), false, -1)
	}

	if profile.IsTweetInDatabase(tweet_id) {
		fmt.Println("Tweet is already in database.  Updating...")
	}
	tweet, err := scraper.GetTweet(tweet_id)
	if err != nil {
		die("Error fetching tweet: " + err.Error(), false, -1)
	}
	fmt.Println(tweet)

	err = profile.SaveTweet(tweet)
	if err != nil {
		die("Error saving tweet: " + err.Error(), false, 4)
	}
	fmt.Println("Saved the tweet.  Exiting successfully")
}


func download_tweet_content(tweet_id string) {
	id, err := strconv.Atoi(tweet_id)
	if err != nil {
		panic(err)
	}
	tweet, err := profile.GetTweetById(scraper.TweetID(id))
	if err != nil {
		panic("Couldn't get tweet from database: " + err.Error())
	}
	err = profile.DownloadTweetContentFor(&tweet)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}

func download_user_content(handle scraper.UserHandle) {
	user, err := profile.GetUserByHandle(handle)
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	err = profile.DownloadUserContentFor(&user)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}
