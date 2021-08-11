package main

import (
	"fmt"
	"os"
	"offline_twitter/scraper"
	"offline_twitter/terminal_utils"
	"strings"
	"strconv"
)


/**
 * Help message to print if command syntax is incorrect
 */
const help_message = `Usage: twitter [--profile <profile_dir>] <operation> <TARGET>

<operation>:
  - create_profile (<TARGET> is the directory to create).
          <TARGET> must not exist.  <profile_dir> will be ignored if provided.

  - fetch_user (<TARGET> is the user handle)
  - fetch_tweet_only (<TARGET> is the full URL of the tweet)
  - download_tweet_content (<TARGET> is the ID of the tweet whomst contents to download / back up)
  - download_user_content (<TARGET> is the user handle of the user whomst banner image and profile to download / back up)

<profile_dir>: the path to the directory containing the data directories, database files, and settings files.  By default, refers to the current directory.  Ignored if <operation> is "create_profile".
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
	parts := strings.Split(url, "/")
	if len(parts) != 6 {
		return 0, fmt.Errorf("Tweet format isn't right (%d)", len(parts))
	}
	if parts[0] != "https:" || parts[1] != "" || parts[2] != "twitter.com" || parts[4] != "status" {
		return 0, fmt.Errorf("Tweet format isn't right")
	}
	id, err := strconv.Atoi(parts[5])
	if err != nil {
		return 0, err
	}
	return scraper.TweetID(id), nil
}
