package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/terminal_utils"
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
func happy_exit(text string, was_rate_limited bool) {
	if was_rate_limited {
		fmt.Printf(terminal_utils.COLOR_YELLOW + text + terminal_utils.COLOR_RESET + "\n")
		fmt.Printf(terminal_utils.COLOR_YELLOW + "Exiting early (rate limited)." + terminal_utils.COLOR_RESET + "\n")
		os.Exit(1)
	}
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

// Get a sensible default path to create a default profile.  Uses `XDG_DATA_HOME` if available
//
// Defaults:
//   - Unix: `~/.local/share`
//   - Windows: %APPDATA%
//   - MacOS:  ~/Library
func get_default_profile() string {
	app_data_dir := os.Getenv("XDG_DATA_HOME")
	if app_data_dir == "" {
		switch runtime.GOOS {
		case "windows":
			app_data_dir = os.Getenv("AppData")
			if app_data_dir == "" {
				panic("%AppData% is undefined")
			}
		case "darwin":
			app_data_dir = filepath.Join(os.Getenv("HOME"), "Library")
		default: // Unix
			app_data_dir = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	}
	return filepath.Join(app_data_dir, "twitter")
}

// Returns whether this error should be treated as a failure
func is_scrape_failure(err error) bool {
	if err == nil || errors.Is(err, scraper.END_OF_FEED) || errors.Is(err, scraper.ErrRateLimited) {
		return false
	}
	return true
}
