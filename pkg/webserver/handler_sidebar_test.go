package webserver_test

import (
	"testing"

	"net/http/httptest"

	"github.com/stretchr/testify/require"
)

func TestSidebar(t *testing.T) {
	require := require.New(t)

	req := httptest.NewRequest("GET", "/nav-sidebar-poll-updates", nil)
	req.Header.Set("HX-Request", "true")
	resp := do_request(req)
	require.Equal(resp.StatusCode, 200)
}
