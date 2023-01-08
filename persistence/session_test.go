package persistence_test

import (
	"github.com/go-test/deep"
	"offline_twitter/persistence"
	"offline_twitter/scraper"
	"testing"
)

// Save and load an API session; it should come back the same
func TestSaveAndLoadAuthenticatedSession(t *testing.T) {
	assert := assert.New(t)
	profile_path := "test_profiles/TestSession"
	profile := create_or_load_profile(profile_path)

	api := scraper.API{
		// TODO session-saving
		// - Fill out some fields here like Cookies and CSRFToken and UserHandle
	}

	// Save and load the session; it should come back the same
	profile.SaveSession(api)
	new_api = profile.LoadSession(api.UserHandle)

	if diff := deep.Equal(api, new_api); diff != nil {
		t.Errorf(diff)
	}
}
