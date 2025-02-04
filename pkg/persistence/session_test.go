package persistence_test

import (
	"fmt"

	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/go-test/deep"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Save and load an API session; it should come back the same
func TestSaveAndLoadAuthenticatedSession(t *testing.T) {
	profile_path := "test_profiles/TestSession"
	profile := create_or_load_profile(profile_path)

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	api := scraper.API{
		UserHandle:      "testUser",
		IsAuthenticated: true,
		Client: http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
		CSRFToken: fmt.Sprint(rand.Int()),
	}

	// Save and load the session; it should come back the same
	profile.SaveSession(api.UserHandle, api.MustMarshalJSON())
	var new_api scraper.API
	profile.LoadSession(api.UserHandle, &new_api)

	if diff := deep.Equal(api, new_api); diff != nil {
		t.Error(diff)
	}
}
