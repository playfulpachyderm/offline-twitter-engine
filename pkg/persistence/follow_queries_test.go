package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestSaveAndLoadFollows(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	follower := create_dummy_user()
	require.NoError(profile.SaveUser(&follower))

	followee_ids := []UserID{
		1427250806378672134,
		1304281147074064385,
		887434912529338375,
		836779281049014272,
		1032468021485293568,
	}
	trove := NewTweetTrove()
	for _, id := range followee_ids {
		trove.Users[id] = User{}
	}

	// Save and reload it
	profile.SaveAsFolloweesList(follower.ID, trove)
	new_followee_ids := profile.GetFollowees(follower.ID)

	assert.Len(new_followee_ids, len(followee_ids))
	for _, id := range new_followee_ids {
		_, is_ok := trove.Users[id]
		assert.True(is_ok)
	}
}

func TestIsFollowing(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	assert.True(profile.IsXFollowingY(UserID(1178839081222115328), UserID(1488963321701171204)))
	assert.False(profile.IsXFollowingY(UserID(1488963321701171204), UserID(1178839081222115328)))
}
