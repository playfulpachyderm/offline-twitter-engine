package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

/**
 * Return the expanded version of a short URL.  Input must be a real short URL.
 */
func ExpandShortUrl(short_url string) string {
	// Create a client that doesn't follow redirects
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(short_url)
	if err != nil {
		panic(err) // TODO: handle timeouts
	}
	if resp.StatusCode != 301 {
		panic(fmt.Errorf("Unknown status code returned when expanding short url %q: %s\n  %w", short_url, resp.Status, EXTERNAL_API_ERROR))
	}

	long_url := resp.Header.Get("Location")
	if long_url == "" {
		panic(fmt.Errorf("Header didn't have a Location field for short url %q:\n  %w", short_url, EXTERNAL_API_ERROR))
	}
	return long_url
}

// Given an URL, try to parse it as a tweet url.
// The bool is an `is_ok` value; true if the parse was successful, false if it didn't match
func TryParseTweetUrl(s string) (UserHandle, TweetID, bool) {
	parsed_url, err := url.Parse(s)
	if err != nil {
		return UserHandle(""), TweetID(0), false
	}

	if parsed_url.Host != "twitter.com" && parsed_url.Host != "mobile.twitter.com" && parsed_url.Host != "x.com" {
		return UserHandle(""), TweetID(0), false
	}

	r := regexp.MustCompile(`^/(\w+)/status/(\d+)$`)
	matches := r.FindStringSubmatch(parsed_url.Path)
	if matches == nil {
		return UserHandle(""), TweetID(0), false
	}
	if len(matches) != 3 { // matches[0] is the full string
		panic(matches)
	}
	return UserHandle(matches[1]), TweetID(int_or_panic(matches[2])), true
}

/**
 * Given a tweet URL, return the corresponding user handle.
 * If tweet url is not valid, return an error.
 */
func ParseHandleFromTweetUrl(tweet_url string) (UserHandle, error) {
	short_url_regex := regexp.MustCompile(`^https://t.co/\w{5,20}$`)
	if short_url_regex.MatchString(tweet_url) {
		tweet_url = ExpandShortUrl(tweet_url)
	}

	ret, _, is_ok := TryParseTweetUrl(tweet_url)
	if !is_ok {
		return "", fmt.Errorf("Invalid tweet url: %s", tweet_url)
	}
	return ret, nil
}
