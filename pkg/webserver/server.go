package webserver

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type Middleware func(http.Handler) http.Handler

type Application struct {
	AccessLog *log.Logger
	TraceLog  *log.Logger
	InfoLog   *log.Logger
	ErrorLog  *log.Logger

	Middlewares []Middleware

	Profile                       Profile
	ActiveUser                    User
	IsScrapingDisabled            bool
	API                           scraper.API
	LastReadNotificationSortIndex int64
}

func NewApp(profile Profile) Application {
	ret := Application{
		AccessLog: log.New(os.Stdout, "ACCESS\t", log.Ldate|log.Ltime),
		TraceLog:  log.New(os.Stdout, "TRACE\t", log.Ldate|log.Ltime),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),

		Profile:            profile,
		ActiveUser:         get_default_user(),
		IsScrapingDisabled: true, // Until an active user is set
	}

	// Can ignore errors; if not authenticated, it won't be used for anything.
	// GetUser and Login both create a new session.
	ret.API, _ = scraper.NewGuestSession() //nolint:errcheck // see above

	ret.Middlewares = []Middleware{
		secureHeaders,
		ret.logRequest,
		ret.recoverPanic,
	}
	return ret
}

func (app *Application) WithMiddlewares() http.Handler {
	var ret http.Handler = app
	for i := range app.Middlewares {
		ret = app.Middlewares[i](ret)
	}
	return ret
}

func (app *Application) SetActiveUser(handle UserHandle) error {
	if handle == "no account" {
		app.ActiveUser = get_default_user()
		app.IsScrapingDisabled = true // API requests will fail b/c not logged in
	} else {
		user, err := app.Profile.GetUserByHandle(handle)
		if err != nil {
			return fmt.Errorf("set active user to %q: %w", handle, err)
		}
		app.Profile.LoadSession(handle, &app.API)
		app.ActiveUser = user
		app.IsScrapingDisabled = false
	}
	return nil
}

func get_default_user() User {
	return User{
		ID:                    0,
		Handle:                "[nobody]",
		DisplayName:           "[Not logged in]",
		ProfileImageLocalPath: path.Base(DEFAULT_PROFILE_IMAGE_URL),
		IsContentDownloaded:   true,
	}
}

// Manual router implementation.
// I don't like the weird matching behavior of http.ServeMux, and it's not hard to write by hand.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/timeline", 303)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[1:]
	switch parts[0] {
	case "static":
		// Static files can be stored in browser cache
		w.Header().Set("Cache-Control", "public, max-age=3600")
		if use_embedded == "true" {
			// Serve directly from the embedded files
			http.FileServer(http.FS(embedded_files)).ServeHTTP(w, r)
		} else {
			// Serve from disk
			http.StripPrefix("/static", http.FileServer(http.Dir(get_filepath("static")))).ServeHTTP(w, r)
		}
	case "tweet":
		app.TweetDetail(w, r)
	case "content":
		http.StripPrefix("/content", http.FileServer(http.Dir(app.Profile.ProfileDir))).ServeHTTP(w, r)
	case "login":
		app.Login(w, r)
	case "change-session":
		app.ChangeSession(w, r)
	case "timeline":
		app.Timeline(w, r)
	case "follow":
		app.UserFollow(w, r)
	case "unfollow":
		app.UserUnfollow(w, r)
	case "search":
		http.StripPrefix("/search", http.HandlerFunc(app.Search)).ServeHTTP(w, r)
	case "lists":
		http.StripPrefix("/lists", http.HandlerFunc(app.Lists)).ServeHTTP(w, r)
	case "bookmarks":
		app.Bookmarks(w, r)
	case "notifications":
		http.StripPrefix("/notifications", http.HandlerFunc(app.Notifications)).ServeHTTP(w, r)
	case "messages":
		http.StripPrefix("/messages", http.HandlerFunc(app.Messages)).ServeHTTP(w, r)
	case "nav-sidebar-poll-updates":
		app.NavSidebarPollUpdates(w, r)
	case "communities":
		panic("not implemented")
	default:
		app.UserFeed(w, r)
	}
}

func (app *Application) Run(address string, should_auto_open bool) {
	srv := &http.Server{
		Addr:     address,
		ErrorLog: app.ErrorLog,
		Handler:  app.WithMiddlewares(),
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.InfoLog.Printf("Starting server on %s", address)

	app.start_background()

	if should_auto_open {
		page := "/login"
		if app.ActiveUser.ID != get_default_user().ID {
			page = "" // Load the timeline
		}
		go openWebPage("http://" + address + page)
	}
	err := srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}

func openWebPage(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to open homepage: %s", err.Error())
	}
}

func parse_cursor_value(c *Cursor, r *http.Request) error {
	cursor_param := r.URL.Query().Get("cursor")
	if cursor_param != "" {
		var err error
		c.CursorValue, err = strconv.Atoi(cursor_param)
		if err != nil {
			return fmt.Errorf("attempted to parse cursor value %q as int: %w", c.CursorValue, err)
		}
		c.CursorPosition = CURSOR_MIDDLE
	}
	return nil
}
