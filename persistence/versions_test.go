package persistence_test

import (
	"testing"
	"os"
	"offline_twitter/scraper"
	"offline_twitter/persistence"
)

func TestVersionUpgrade(t *testing.T) {
	profile_path := "test_profiles/TestVersions"
	if file_exists(profile_path) {
		err := os.RemoveAll(profile_path)
		if err != nil {
			panic(err)
		}
	}
	profile := create_or_load_profile(profile_path)

	test_migration := "insert into tweets (id, user_id, text) values (21250554358298342, -1, 'awefjk')"
	test_tweet_id := scraper.TweetID(21250554358298342)

	if profile.IsTweetInDatabase(test_tweet_id) {
		t.Fatalf("Test tweet shouldn't be in the database yet")
	}

	persistence.MIGRATIONS = append(persistence.MIGRATIONS, test_migration)
	profile.UpgradeFromXToY(persistence.ENGINE_DATABASE_VERSION, persistence.ENGINE_DATABASE_VERSION + 1)

	if !profile.IsTweetInDatabase(test_tweet_id) {
		t.Errorf("Migration should have created the tweet, but it didn't")
	}
}
