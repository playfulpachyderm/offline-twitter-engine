package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"
)

func TestSaveAndLoadLike(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestLikesQueries"
	profile := create_or_load_profile(profile_path)

	like := create_dummy_like()
	err := profile.SaveLike(like)
	require.NoError(err)

	// Reload the Like
	new_like, err := profile.GetLikeBySortID(like.SortID)
	require.NoError(err)

	// Should come back the same
	if diff := deep.Equal(like, new_like); diff != nil {
		t.Error(diff)
	}

	// Test double-saving
	err = profile.SaveLike(like)
	require.NoError(err)
	new_like, err = profile.GetLikeBySortID(like.SortID)
	require.NoError(err)
	if diff := deep.Equal(like, new_like); diff != nil {
		t.Error(diff)
	}
}

func TestDeleteLike(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestLikesQueries"
	profile := create_or_load_profile(profile_path)

	like := create_dummy_like()
	err := profile.SaveLike(like)
	require.NoError(err)

	// Delete it
	err = profile.DeleteLike(like)
	require.NoError(err)

	// Should be gone
	_, err = profile.GetLikeBySortID(like.SortID)
	require.Error(err)
}
