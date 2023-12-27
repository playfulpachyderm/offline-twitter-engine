package scraper_test

import (
	"testing"

	"encoding/json"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseFolloweesList(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/lists/followees.json")
	require.NoError(err)

	var resp APIV2Response
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	tweet_trove, err := resp.ToTweetTrove()
	require.NoError(err)

	// Check users
	assert.Len(tweet_trove.Users, 4)
	_, is_ok := tweet_trove.Users[1349149096909668363]
	assert.True(is_ok)

	// Test cursor-bottom
	bottom_cursor := resp.GetCursorBottom()
	assert.Equal("0|1739810405452087290", bottom_cursor)
}
