package persistence_test

import (
	"testing"

	"os"

	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestVersionUpgrade(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestVersions"
	if file_exists(profile_path) {
		err := os.RemoveAll(profile_path)
		require.NoError(err)
	}
	profile := create_or_load_profile(profile_path)

	test_migration := "insert into tweets (id, user_id, text) values (21250554358298342, -1, 'awefjk')"
	test_tweet_id := TweetID(21250554358298342)

	require.False(profile.IsTweetInDatabase(test_tweet_id), "Test tweet shouldn't be in db yet")

	MIGRATIONS = append(MIGRATIONS, test_migration)
	err := profile.UpgradeFromXToY(ENGINE_DATABASE_VERSION, ENGINE_DATABASE_VERSION+1)
	require.NoError(err)

	require.True(profile.IsTweetInDatabase(test_tweet_id), "Migration should have created the tweet, but it didn't")
}
