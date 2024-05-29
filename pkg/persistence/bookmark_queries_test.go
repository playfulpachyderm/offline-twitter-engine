package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"
)

func TestSaveAndLoadBookmark(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestBookmarksQueries"
	profile := create_or_load_profile(profile_path)

	bookmark := create_dummy_bookmark()
	err := profile.SaveBookmark(bookmark)
	require.NoError(err)

	// Reload the Bookmark
	new_bookmark, err := profile.GetBookmarkBySortID(bookmark.SortID)
	require.NoError(err)

	// Should come back the same
	if diff := deep.Equal(bookmark, new_bookmark); diff != nil {
		t.Error(diff)
	}

	// Test double-saving
	err = profile.SaveBookmark(bookmark)
	require.NoError(err)
	new_bookmark, err = profile.GetBookmarkBySortID(bookmark.SortID)
	require.NoError(err)
	if diff := deep.Equal(bookmark, new_bookmark); diff != nil {
		t.Error(diff)
	}
}

func TestDeleteBookmark(t *testing.T) {
	require := require.New(t)

	profile_path := "test_profiles/TestBookmarksQueries"
	profile := create_or_load_profile(profile_path)

	bookmark := create_dummy_bookmark()
	err := profile.SaveBookmark(bookmark)
	require.NoError(err)

	// Delete it
	err = profile.DeleteBookmark(bookmark)
	require.NoError(err)

	// Should be gone
	_, err = profile.GetBookmarkBySortID(bookmark.SortID)
	require.Error(err)
}
