package webserver_test

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/andybalholm/cascadia"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/webserver"
)

type CapturingWriter struct {
	Writes [][]byte
}

func (w *CapturingWriter) Write(p []byte) (int, error) {
	w.Writes = append(w.Writes, p)
	return len(p), nil
}

var profile Profile

func init() {
	var err error
	profile, err = LoadProfile("../../sample_data/profile")
	if err != nil {
		panic(err)
	}
}

func selector(s string) cascadia.Sel {
	ret, err := cascadia.Parse(s)
	if err != nil {
		panic(err)
	}
	return ret
}

// Run an HTTP request against the app and return the response
func do_request(req *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.WithMiddlewares().ServeHTTP(recorder, req)
	return recorder.Result()
}

// Run an HTTP request against the app, with an Active User set, and return the response
func do_request_with_active_user(req *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	app := webserver.NewApp(profile)
	app.IsScrapingDisabled = true
	app.ActiveUser = User{ID: 1488963321701171204, Handle: "Offline_Twatter"} // Simulate a login
	app.WithMiddlewares().ServeHTTP(recorder, req)
	return recorder.Result()
}

// Homepage
// --------

// Should redirect to the timeline
func TestHomepage(t *testing.T) {
	require := require.New(t)

	resp := do_request(httptest.NewRequest("GET", "/", nil))
	require.Equal(resp.StatusCode, 303)
	require.Equal(resp.Header.Get("Location"), "/timeline")
}
