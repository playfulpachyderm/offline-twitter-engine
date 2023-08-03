package persistence_test

import (
	"testing"

	"errors"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// DUPE 1
func file_exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

// Should refuse to create a Profile if the target already exists (i.e., is a file or directory).
func TestNewProfileInvalidPath(t *testing.T) {
	require := require.New(t)
	gibberish_path := "test_profiles/fjlwrefuvaaw23efwm"
	if file_exists(gibberish_path) {
		err := os.RemoveAll(gibberish_path)
		require.NoError(err)
	}
	err := os.Mkdir(gibberish_path, 0755)
	require.NoError(err)

	_, err = persistence.NewProfile(gibberish_path)
	require.Error(err, "Should have failed to create a profile in an already existing directory!")

	assert.ErrorIs(t, err, persistence.ErrTargetAlreadyExists)
}

// Should correctly create a new Profile
func TestNewProfile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile_path := "test_profiles/TestNewProfile"
	if file_exists(profile_path) {
		err := os.RemoveAll(profile_path)
		require.NoError(err)
	}

	profile, err := persistence.NewProfile(profile_path)
	require.NoError(err)

	assert.Equal(profile_path, profile.ProfileDir)

	// Check files were created
	contents, err := os.ReadDir(profile_path)
	require.NoError(err)
	assert.Len(contents, 7)

	expected_files := []struct {
		filename string
		isDir    bool
	}{
		{"images", true},
		{"link_preview_images", true},
		{"profile_images", true},
		{"settings.yaml", false},
		{"twitter.db", false},
		{"video_thumbnails", true},
		{"videos", true},
	}

	for i, v := range expected_files {
		assert.Equal(v.filename, contents[i].Name())
		assert.Equal(v.isDir, contents[i].IsDir())
	}

	// Check database version is initialized
	version, err := profile.GetDatabaseVersion()
	require.NoError(err)
	assert.Equal(persistence.ENGINE_DATABASE_VERSION, version)
}

// Should correctly load the Profile
func TestLoadProfile(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestLoadProfile"
	if file_exists(profile_path) {
		err := os.RemoveAll(profile_path)
		require.NoError(err)
	}

	_, err := persistence.NewProfile(profile_path)
	require.NoError(err)

	profile, err := persistence.LoadProfile(profile_path)
	require.NoError(err)

	assert.Equal(t, profile_path, profile.ProfileDir)
}

// Test the ToSnakeCase implementation
func TestSnakeCase(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("tweet_id", persistence.ToSnakeCase("TweetID"))
	assert.Equal("i_am_a_computer", persistence.ToSnakeCase("IAmAComputer"))
	assert.Equal("choice1_votes", persistence.ToSnakeCase("Choice1_Votes"))
}
