package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"offline_twitter/scraper"

	"github.com/go-test/deep"
	"math/rand"
)

// Create a Space, save it, reload it, and make sure it comes back the same
func TestSaveAndLoadSpace(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	space := create_space_from_id(rand.Int())
	err := profile.SaveSpace(space)
	require.NoError(err)

	new_space, err := profile.GetSpaceById(space.ID)
	require.NoError(err)
	if diff := deep.Equal(space, new_space); diff != nil {
		t.Error(diff)
	}
}

// Should update a Space
func TestModifySpace(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	space := create_space_from_id(rand.Int())
	err := profile.SaveSpace(space)
	require.NoError(err)

	// Modify and save
	space.State = "Some other state"
	space.UpdatedAt = scraper.TimestampFromUnix(9001)
	space.EndedAt = scraper.TimestampFromUnix(10001)
	space.ReplayWatchCount = 100
	space.LiveListenersCount = 50
	space.IsDetailsFetched = true
	err = profile.SaveSpace(space)
	require.NoError(err)

	new_space, err := profile.GetSpaceById(space.ID)
	require.NoError(err)
	assert.Equal(scraper.TimestampFromUnix(9001), new_space.UpdatedAt)
	assert.Equal(scraper.TimestampFromUnix(10001), new_space.EndedAt)
	assert.Equal(100, new_space.ReplayWatchCount)
	assert.Equal(50, new_space.LiveListenersCount)
	assert.True(new_space.IsDetailsFetched)
}

func TestNoWorseningSpace(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	profile_path := "test_profiles/TestMediaQueries"
	profile := create_or_load_profile(profile_path)

	space := create_space_from_id(rand.Int())
	space.ShortUrl = "Some Short Url"
	space.State = "Some State"
	space.Title = "Debating Somebody"
	space.CreatedAt = scraper.TimestampFromUnix(1000)
	space.UpdatedAt = scraper.TimestampFromUnix(2000)
	space.CreatedById = scraper.UserID(-1)
	space.LiveListenersCount = 100
	space.IsDetailsFetched = true

	// Save the space
	err := profile.SaveSpace(space)
	require.NoError(err)

	// Worsen the space, then re-save
	space.ShortUrl = ""
	space.Title = ""
	space.State = ""
	space.CreatedAt = scraper.TimestampFromUnix(0)
	space.UpdatedAt = scraper.TimestampFromUnix(0)
	space.CreatedById = scraper.UserID(0)
	space.LiveListenersCount = 0
	space.IsDetailsFetched = false
	err = profile.SaveSpace(space)
	require.NoError(err)

	// Reload it
	new_space, err := profile.GetSpaceById(space.ID)
	require.NoError(err)

	assert.Equal(new_space.ShortUrl, "Some Short Url")
	assert.Equal(new_space.State, "Some State")
	assert.Equal(new_space.Title, "Debating Somebody")
	assert.Equal(new_space.CreatedAt, scraper.TimestampFromUnix(1000))
	assert.Equal(new_space.UpdatedAt, scraper.TimestampFromUnix(2000))
	assert.Equal(new_space.CreatedById, scraper.UserID(-1))
	assert.Equal(new_space.LiveListenersCount, 100)
	assert.True(new_space.IsDetailsFetched)
}
