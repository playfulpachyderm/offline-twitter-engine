package scraper_test

import (
	. "offline_twitter/scraper"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthentication(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	username := "offline_twatter"
	password := "S1pKIW#eRT016iA@OFcK"

	api := NewGuestSession()
	api.LogIn(username, password)

	assert.True(api.IsAuthenticated)
	assert.NotEqual(api.CSRFToken, "")
	assert.Equal(api.UserHandle, UserHandle("Offline_Twatter"))

	response, err := api.GetLikesFor(1458284524761075714, "")
	require.NoError(err)
	trove, err := response.ToTweetTrove()
	require.NoError(err)
	assert.True(len(trove.Tweets) > 0)
}
