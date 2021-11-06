package main

import (
	"fmt"
	"os"
	"offline_twitter/scraper"
	"offline_twitter/terminal_utils"
	"strings"
	"strconv"
	"regexp"
)


/**
 * Help message to print if command syntax is incorrect
 */
const help_message = `Usage: twitter [--profile <profile_dir>] <operation> <TARGET>
This application downloads tweets from twitter and saves them in a SQLite database.

<profile_dir>:
    Optional.  Indicates the path to the directory containing the data directories, database files, and settings files.
    By default, will use the current working directory.
    Ignored if <operation> is "create_profile".

<operation>:
    create_profile
          <TARGET> is the directory to create.  It must not exist already.
          <profile_dir> will be ignored if provided.

    fetch_user
    download_user_content
          <TARGET> is the user handle.
          "download_user_content" will save a local copy of the user's banner and profile images.

    fetch_tweet
    fetch_tweet_only
          <TARGET> is either the full URL of the tweet, or its ID.
          If using "fetch_tweet_only", then only that specific tweet will be saved.  "fetch_tweet" will save the whole thread including replies.

    download_tweet_content
          <TARGET> is either the full URL of the tweet, or its ID.
          Downloads videos and images embedded in the tweet.

    get_user_tweets
    get_user_tweets_all
          <TARGET> is the user handle.
          Gets the most recent ~50 tweets.
          If "get_user_tweets_all" is used, gets up to ~3200 tweets (API limit).

    search
          <TARGET> is the search query.  Should be wrapped in quotes if it has spaces.
`


/**
 * Helper function
 */
func die(text string, display_help bool, exit_code int) {
	if text != "" {
		fmt.Fprint(os.Stderr, terminal_utils.COLOR_RED + text + terminal_utils.COLOR_RESET + "\n")
	}
	if display_help {
		fmt.Fprint(os.Stderr, help_message)
	}
	os.Exit(exit_code)
}

/**
 * Helper function - parse a tweet permalink URL to extract the tweet ID
 *
 * args:
 * - url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 *
 * returns: the id at the end of the tweet: e.g., 1395882872729477131
 */
func extract_id_from(url string) (scraper.TweetID, error) {
	var id_str string

	if regexp.MustCompile(`^\d+$`).MatchString(url) {
		id_str = url
	} else {
		parts := strings.Split(url, "/")
		if len(parts) != 6 {
			return 0, fmt.Errorf("Tweet format isn't right (%d)", len(parts))
		}
		if parts[0] != "https:" || parts[1] != "" || parts[2] != "twitter.com" || parts[4] != "status" {
			return 0, fmt.Errorf("Tweet format isn't right")
		}
		id_str = parts[5]
	}
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return 0, err
	}
	return scraper.TweetID(id), nil
}
