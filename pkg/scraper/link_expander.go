package scraper

import (
	"fmt"
	"net/http"
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
