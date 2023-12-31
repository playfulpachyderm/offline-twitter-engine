package scraper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Makes an HTTP request
func TestGetGuestToken(t *testing.T) {
	token, err := GetGuestToken()
	require.NoError(t, err)

	assert.True(t, len(token) >= 15, "I don't think this is a token: %q", token)
	fmt.Println(token)
}

// Tests the caching.  Should run much much faster than an HTTP request, since all requests
// other than the first use the cache.
func BenchmarkGetGuestToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetGuestToken() //nolint:errcheck  // Don't care about errors, just want to time it
	}
}
