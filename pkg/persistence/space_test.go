package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestFormatSpaceDuration(t *testing.T) {
	assert := assert.New(t)
	s := Space{
		StartedAt: TimestampFromUnix(1000),
		EndedAt:   TimestampFromUnix(5000),
	}
	assert.Equal(s.FormatDuration(), "1h06m")

	s.EndedAt = TimestampFromUnix(500000)
	assert.Equal(s.FormatDuration(), "138h36m")

	s.EndedAt = TimestampFromUnix(1005)
	assert.Equal(s.FormatDuration(), "0m05s")
}
