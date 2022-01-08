package scraper

import (
	"fmt"
	"time"
	"net/http"
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
		panic(err)  // TODO: handle timeouts
	}
	if resp.StatusCode != 301 {
		panic(fmt.Sprintf("Unknown status code returned when expanding short url %q: %s", short_url, resp.Status))
	}

	long_url := resp.Header.Get("Location")
	if long_url == "" {
		panic(fmt.Sprintf("Header didn't have a Location field for short url %q", short_url))
	}
	return long_url
}
