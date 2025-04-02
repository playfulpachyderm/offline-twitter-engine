package webserver_test

import (
	"testing"

	"net/http/httptest"
	"github.com/stretchr/testify/require"
)

func TestFollowUnfollowPostOnly(t *testing.T) {
	require := require.New(t)
	resp := do_request(httptest.NewRequest("GET", "/follow/kwamurai", nil))
	require.Equal(resp.StatusCode, 405)
	resp = do_request(httptest.NewRequest("GET", "/unfollow/kwamurai", nil))
	require.Equal(resp.StatusCode, 405)
}
