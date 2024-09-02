package persistence_test

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
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

func TestGetUnreadNotificationsCount(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	unread_notifs_count := profile.GetUnreadNotificationsCount(1724372973735)
	assert.Equal(2, unread_notifs_count)
}
