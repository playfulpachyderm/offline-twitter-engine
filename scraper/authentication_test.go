package scraper_test

import (
	"offline_twitter/scraper"
	"testing"
)

func TestAuthentication(t *testing.T) {
	username := "offline_twatter"
	password := "S1pKIW#eRT016iA@OFcK"

	api := scraper.NewGuestSession()
	api.LogIn(username, password)
}
