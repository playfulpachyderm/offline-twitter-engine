package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"offline_twitter/persistence"
	"offline_twitter/scraper"
)

/**
 * Global variable referencing the open data profile
 */
var profile persistence.Profile

var version_string string

/**
 * Main method
 */
func main() {
	profile_dir := flag.String("profile", ".", "")
	flag.StringVar(profile_dir, "p", ".", "")

	show_version_flag := flag.Bool("version", false, "")
	flag.BoolVar(show_version_flag, "v", false, "")

	how_many := flag.Int("n", 50, "")
	flag.IntVar(how_many, "number", 50, "")

	var default_log_level string
	if version_string == "" {
		default_log_level = "debug"
	} else {
		default_log_level = "info"
	}
	log_level := flag.String("log-level", default_log_level, "")

	help := flag.Bool("help", false, "")
	flag.BoolVar(help, "h", false, "")

	flag.Usage = func() {
		die("", true, 1)
	}

	flag.Parse()
	args := flag.Args()

	if *show_version_flag {
		if version_string == "" {
			fmt.Println("Development version")
		} else {
			fmt.Println("v" + version_string)
		}
		os.Exit(0)
	}

	if *help {
		die("", true, 0)
	}

	logging_level, err := log.ParseLevel(*log_level)
	if err != nil {
		die(err.Error(), false, 1)
	}
	log.SetLevel(logging_level)

	if len(args) < 2 {
		if len(args) == 1 && args[0] == "list_followed" {
			// "list_followed" doesn't need a target, so create a fake second arg
			args = append(args, "")
		} else {
			die("", true, 1)
		}
	}

	operation := args[0]
	target := args[1]

	if operation == "create_profile" {
		create_profile(target)
		return
	}

	profile, err = persistence.LoadProfile(*profile_dir)
	if err != nil {
		die(fmt.Sprintf("Could not load profile: %s", err.Error()), true, 2)
	}

	switch operation {
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
		fetch_user_feed(target, *how_many)
	case "get_user_tweets_all":
		fetch_user_feed(target, 999999999)
	case "download_tweet_content":
		download_tweet_content(target)
	case "search":
		search(target)
	case "follow":
		follow_user(target, true)
	case "unfollow":
		follow_user(target, false)
	case "list_followed":
		list_followed()
	default:
		die(fmt.Sprintf("Invalid operation: %s", operation), true, 3)
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
	user, err := scraper.GetUser(handle)
	if err != nil {
		die(err.Error(), false, -1)
	}
	log.Debug(user)

	err = profile.SaveUser(&user)
	if err != nil {
		die(fmt.Sprintf("Error saving user: %s", err.Error()), false, 4)
	}

	download_user_content(handle)
	happy_exit("Saved the user")
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

	tweet, err := scraper.GetTweet(tweet_id)
	if err != nil {
		die(fmt.Sprintf("Error fetching tweet: %s", err.Error()), false, -1)
	}
	log.Debug(tweet)

	err = profile.SaveTweet(tweet)
	if err != nil {
		die(fmt.Sprintf("Error saving tweet: %s", err.Error()), false, 4)
	}
	happy_exit("Saved the tweet")
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

	trove, err := scraper.GetTweetFull(tweet_id)
	if err != nil {
		die(err.Error(), false, -1)
	}
	profile.SaveTweetTrove(trove)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)))
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
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := scraper.GetUserFeedGraphqlFor(user.ID, how_many)
	if err != nil {
		die(fmt.Sprintf("Error scraping feed: %s\n  %s", handle, err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove)

	happy_exit(fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)))
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
	trove, err := scraper.Search(query, 1000)
	if err != nil {
		die(fmt.Sprintf("Error scraping search results: %s", err.Error()), false, -100)
	}
	profile.SaveTweetTrove(trove)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)))
}

func follow_user(handle string, is_followed bool) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	profile.SetUserFollowed(&user, is_followed)

	if is_followed {
		happy_exit("Followed user: " + handle)
	} else {
		happy_exit("Unfollowed user: " + handle)
	}
}

func list_followed() {
	for _, handle := range profile.GetAllFollowedUsers() {
		fmt.Println(handle)
	}
}
