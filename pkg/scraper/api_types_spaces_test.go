package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseSpaceResponse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/space_object.json")
	if err != nil {
		panic(err)
	}

	var response SpaceResponse
	err = json.Unmarshal(data, &response)
	assert.NoError(err)

	trove := response.ToTweetTrove()
	require.Len(trove.Spaces, 1)
	space := trove.Spaces["1BdxYypQzBgxX"]
	assert.Equal(space.Title, "dreary weather 🌧️☔🌬️")
	assert.Equal(space.CreatedById, UserID(1356335022815539201))
	assert.Equal(int64(1665884387), space.CreatedAt.Time.Unix())
	assert.Equal(int64(1665884388), space.StartedAt.Time.Unix())
	assert.Equal(int64(1665887491), space.EndedAt.Time.Unix())
	assert.Equal(int64(1665887492), space.UpdatedAt.Time.Unix())
	assert.False(space.IsAvailableForReplay)
	assert.Equal(4, space.ReplayWatchCount)
	assert.Equal(1, space.LiveListenersCount)

	assert.True(space.IsDetailsFetched)

	assert.Len(space.ParticipantIds, 2)
	assert.Equal(UserID(1356335022815539201), space.ParticipantIds[0])
	assert.Equal(UserID(1523838615377350656), space.ParticipantIds[1])

	require.Len(trove.Users, 1)
	user := trove.Users[1356335022815539201]
	assert.Equal(847, user.FollowersCount)
}

func TestParseEmptySpaceResponse(t *testing.T) {
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/space_object_empty.json")
	if err != nil {
		panic(err)
	}

	var response SpaceResponse
	err = json.Unmarshal(data, &response)
	require.NoError(err)

	trove := response.ToTweetTrove()
	require.Len(trove.Spaces, 0)
}
