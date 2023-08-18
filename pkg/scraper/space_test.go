package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseSpace(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/space.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	space := ParseAPISpace(apiCard)
	assert.Equal(SpaceID("1YpKkZVyQjoxj"), space.ID)
	assert.Equal("https://t.co/WBPAHNF8Om", space.ShortUrl)
}

func TestFormatSpaceDuration(t *testing.T) {
	assert := assert.New(t)
	s := Space{
		StartedAt: TimestampFromUnix(1000 * 1000),
		EndedAt:   TimestampFromUnix(5000 * 1000),
	}
	assert.Equal(s.FormatDuration(), "1h06m")

	s.EndedAt = TimestampFromUnix(500000 * 1000)
	assert.Equal(s.FormatDuration(), "138h36m")

	s.EndedAt = TimestampFromUnix(1005 * 1000)
	assert.Equal(s.FormatDuration(), "0m05s")
}
