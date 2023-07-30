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
