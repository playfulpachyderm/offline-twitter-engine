package main

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"

	"offline_twitter/scraper"
	"offline_twitter/terminal_utils"
)

/**
 * Help message to print if command syntax is incorrect
 */
//go:embed help_message.txt
var help_message string

/**
 * Helper function
 */
func die(text string, display_help bool, exit_code int) {
	if text != "" {
		outstring := terminal_utils.COLOR_RED + text + terminal_utils.COLOR_RESET + "\n"
		fmt.Fprint(os.Stderr, outstring)
	}
	if display_help {
		fmt.Fprint(os.Stderr, help_message)
	}
	os.Exit(exit_code)
}

/**
 * Print a happy exit message and exit
 */
func happy_exit(text string) {
	fmt.Printf(terminal_utils.COLOR_GREEN + text + terminal_utils.COLOR_RESET + "\n")
	fmt.Printf(terminal_utils.COLOR_GREEN + "Exiting successfully." + terminal_utils.COLOR_RESET + "\n")
	os.Exit(0)
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
	_, id, is_ok := scraper.TryParseTweetUrl(url)
	if is_ok {
		return id, nil
	}

	num, err := strconv.Atoi(url)
	return scraper.TweetID(num), err
}
