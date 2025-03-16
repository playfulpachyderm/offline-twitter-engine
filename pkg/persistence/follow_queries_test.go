package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// Follow a bunch of users, then check that they are saved properly
func TestSaveAndLoadFollows(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
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
	new_followees := profile.GetFollowees(follower.ID)

	assert.Len(new_followees, len(followee_ids))
	for _, followee := range new_followees {
		_, is_ok := trove.Users[followee.ID]
		assert.True(is_ok)
	}
}

func TestIsFollowing(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	assert.True(profile.IsXFollowingY(UserID(1178839081222115328), UserID(1488963321701171204)))
	assert.False(profile.IsXFollowingY(UserID(1488963321701171204), UserID(1178839081222115328)))
}

func TestFollowUnfollow(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile_path := "test_profiles/TestUserQueries"
	profile := create_or_load_profile(profile_path)

	u1 := create_stable_user()
	u2 := create_dummy_user()
	require.NoError(profile.SaveUser(&u2))

	// Should not be following to begin
	assert.False(profile.IsXFollowingY(u1.ID, u2.ID))

	// Follow them
	profile.SaveFollow(u1.ID, u2.ID)
	assert.True(profile.IsXFollowingY(u1.ID, u2.ID))

	// Unfollow-- should be gone
	profile.DeleteFollow(u1.ID, u2.ID)
	assert.False(profile.IsXFollowingY(u1.ID, u2.ID))
}
