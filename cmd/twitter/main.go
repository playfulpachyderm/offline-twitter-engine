package main

import (
	"os"
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
		die("", true, 0)
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
	case "download_user_content":
		download_user_content(scraper.UserHandle(target))
	case "fetch_tweet_only":
		fetch_tweet_only(target)
	case "fetch_tweet":
		fetch_tweet_conversation(target)
	case "get_user_tweets":
		fetch_user_feed(target, 50)
	case "get_user_tweets_all":
		fetch_user_feed(target, 999999999)
	case "download_tweet_content":
		download_tweet_content(target)
	case "search":
		search(target)
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

	fmt.Println("Saved the user.  Downloading content..")

	download_user_content(handle);
}

/**
 * Scrape a single tweet and save it in the database.
 *
 * args:
 * - tweet_url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 */
func fetch_tweet_only(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
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

/**
 * Scrape a tweet and all associated info, and save it in the database.
 *
 * args:
 * - tweet_url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 */
func fetch_tweet_conversation(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}

	if profile.IsTweetInDatabase(tweet_id) {
		fmt.Println("Tweet is already in database.  Updating...")
	}

	tweets, _, users, err := scraper.GetTweetFull(tweet_id)
	if err != nil {
		die(err.Error(), false, -1)
	}

	for _, u := range users {
		fmt.Println(u.Handle)
		_, err = profile.DownloadUserContentIfNeeded(&u)
		if err != nil {
			die("Error getting user content: " + err.Error(), false, 10)
		}

		err = profile.SaveUser(u)
		if err != nil {
			die("Error saving user: " + err.Error(), false, 4)
		}
	}

	for _, t := range tweets {
		err = profile.SaveTweet(t)
		if err != nil {
			die(fmt.Sprintf("Error saving tweet (id %d): %s", t.ID, err.Error()), false, 4)
		}
		err = profile.DownloadTweetContentFor(&t)
		if err != nil {
			die("Error getting tweet content: " + err.Error(), false, 11)
		}
	}
	fmt.Printf("Saved %d tweets and %d users.  Exiting successfully\n", len(tweets), len(users))
}

/**
 * Scrape a user feed and get a big blob of tweets and retweets.  Get 50 tweets.
 *
 * args:
 * - handle: the user handle to get
 */
func fetch_user_feed(handle string, how_many int) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))

	if err != nil {
		die(err.Error(), false, -1)
	}

	tweets, retweets, users, err := scraper.GetUserFeedFor(user.ID, how_many);
	if err != nil {
		die("Error scraping feed: " + err.Error(), false, -2)
	}

	for _, u := range users {
		fmt.Println(u.Handle)
		_, err = profile.DownloadUserContentIfNeeded(&u)
		if err != nil {
			die("Error getting user content: " + err.Error(), false, 10)
		}
		err = profile.SaveUser(u)
		if err != nil {
			die("Error saving user: " + err.Error(), false, 4)
		}
	}

	for _, t := range tweets {
		err = profile.SaveTweet(t)
		if err != nil {
			die("Error saving tweet: " + err.Error(), false, 4)
		}
		err = profile.DownloadTweetContentFor(&t)
		if err != nil {
			die("Error getting tweet content: " + err.Error(), false, 11)
		}
	}

	for _, r := range retweets {
		err = profile.SaveRetweet(r)
		if err != nil {
			die("Error saving retweet: " + err.Error(), false, 4)
		}
	}

	fmt.Printf("Saved %d tweets, %d retweets and %d users.  Exiting successfully\n", len(tweets), len(retweets), len(users))
}


func download_tweet_content(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}

	tweet, err := profile.GetTweetById(tweet_id)
	if err != nil {
		panic(fmt.Sprintf("Couldn't get tweet (ID %d) from database: %s", tweet_id, err.Error()))
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


func search(query string) {
	tweets, retweets, users, err := scraper.Search(query, 1000);
	if err != nil {
		die("Error scraping search results: " + err.Error(), false, -100)
	}

	for _, u := range users {
		err = profile.SaveUser(u)
		if err != nil {
			die("Error saving user: " + err.Error(), false, 4)
		}
		err = profile.DownloadUserContentFor(&u)
		if err != nil {
			die("Error getting user content: " + err.Error(), false, 10)
		}
	}

	for _, t := range tweets {
		// fmt.Println(t)
		err = profile.SaveTweet(t)
		if err != nil {
			die("Error saving tweet: " + err.Error(), false, 4)
		}
		err = profile.DownloadTweetContentFor(&t)
		if err != nil {
			die("Error getting tweet content: " + err.Error(), false, 11)
		}
	}

	fmt.Printf("Saved %d tweets, %d retweets and %d users.  Exiting successfully\n", len(tweets), len(retweets), len(users))
}
