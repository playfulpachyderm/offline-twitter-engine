package main

import (
	"fmt"
	"os"
	"offline_twitter/scraper"
	"offline_twitter/terminal_utils"
	"strings"
)


/**
 * Help message to print if command syntax is incorrect
 */
const help_message = `Usage: twitter <operation> <profile_dir> [TARGET]

<operation>:
  - create_profile (no target needed)

  - fetch_user (TARGET is the user handle)
  - fetch_tweet (TARGET is the full URL of the tweet)
  - fetch_tweet_and_replies (TARGET is the full URL of the tweet)

<profile_dir>: the path to the directory containing the data directories, database files, and settings files.

TARGET is optional depending on <operation>
`


/**
 * Helper function
 */
func die(text string, display_help bool, exit_code int) {
	if text != "" {
		fmt.Print(terminal_utils.COLOR_RED + text + terminal_utils.COLOR_RESET + "\n")
	}
	if display_help {
		fmt.Print(help_message)
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
	parts := strings.Split(url, "/")
	if len(parts) != 6 {
		return "", fmt.Errorf("Tweet format isn't right (%d)", len(parts))
	}
	if parts[0] != "https:" || parts[1] != "" || parts[2] != "twitter.com" || parts[4] != "status" {
		return "", fmt.Errorf("Tweet format isn't right")
	}
	return scraper.TweetID(parts[5]), nil
}
