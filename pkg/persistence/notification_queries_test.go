package persistence_test

import (
	"testing"

	"github.com/go-test/deep"
)

func TestSaveAndLoadNotification(t *testing.T) {
	profile_path := "test_profiles/TestNotificationQuery"
	profile := create_or_load_profile(profile_path)

	// Save it
	n := create_dummy_notification()
	profile.SaveNotification(n)

	// Check it comes back the same
	new_n := profile.GetNotification(n.ID)
	if diff := deep.Equal(n, new_n); diff != nil {
		t.Error(diff)
	}
}
