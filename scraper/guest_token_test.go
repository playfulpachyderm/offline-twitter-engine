package scraper_test

import (
	"testing"
	"fmt"
	. "offline_twitter/scraper"
)

// Makes an HTTP request
func TestGetGuestToken(t *testing.T) {
	token, err := GetGuestToken()
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(token) < 15 {
		t.Errorf("I don't think this is a token: %q", token)
	}
	fmt.Println(token)
}


// Tests the caching.  Should run much much faster than an HTTP request, since all requests
// other than the first use the cache.
func BenchmarkGetGuestToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetGuestToken()
	}
}
