package scraper_test

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestExpandShortUrl(t *testing.T) {
	redirecting_to := "redirect target"
	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Location", redirecting_to)
		w.WriteHeader(301)
	}))
	defer srvr.Close()

	assert.Equal(t, redirecting_to, ExpandShortUrl(srvr.URL))
}
