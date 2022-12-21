package scraper_test

import (
	"offline_twitter/scraper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthentication(t *testing.T) {
	assert := assert.New(t)

	username := "offline_twatter"
	password := "S1pKIW#eRT016iA@OFcK"

	api := scraper.NewGuestSession()
	api.LogIn(username, password)

	assert.True(api.IsAuthenticated)
	assert.NotEqual(api.CSRFToken, "")

}
