package webserver

import (
	"crypto/tls"
	// "encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/form/v4"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type Middleware func(http.Handler) http.Handler

type Application struct {
	accessLog *log.Logger
	traceLog  *log.Logger
	InfoLog   *log.Logger
	ErrorLog  *log.Logger

	Middlewares []Middleware

	Profile            persistence.Profile
	ActiveUser         scraper.User
	IsScrapingDisabled bool
}

func NewApp(profile persistence.Profile) Application {
	ret := Application{
		accessLog: log.New(os.Stdout, "ACCESS\t", log.Ldate|log.Ltime),
		traceLog:  log.New(os.Stdout, "TRACE\t", log.Ldate|log.Ltime),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),

		Profile:            profile,
		ActiveUser:         get_default_user(),
		IsScrapingDisabled: true, // Until an active user is set
	}
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

func (app *Application) SetActiveUser(handle scraper.UserHandle) error {
	if handle == "no account" {
		scraper.InitApi(scraper.NewGuestSession())
		app.ActiveUser = get_default_user()
		app.IsScrapingDisabled = true // API requests will fail b/c not logged in
	} else {
		user, err := app.Profile.GetUserByHandle(handle)
		if err != nil {
			return fmt.Errorf("set active user to %q: %w", handle, err)
		}
		scraper.InitApi(app.Profile.LoadSession(handle))
		app.ActiveUser = user
		app.IsScrapingDisabled = false
	}
	return nil
}

func get_default_user() scraper.User {
	return scraper.User{
		ID:                    0,
		Handle:                "[nobody]",
		DisplayName:           "[Not logged in]",
		ProfileImageLocalPath: path.Base(scraper.DEFAULT_PROFILE_IMAGE_URL),
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
		if use_embedded == "true" {
			// Serve directly from the embedded files
			http.FileServer(http.FS(embedded_files)).ServeHTTP(w, r)
		} else {
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
		app.Lists(w, r)
	default:
		app.UserFeed(w, r)
	}
}

func (app *Application) Run(address string) {
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

	err := srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}

func parse_cursor_value(c *persistence.Cursor, r *http.Request) error {
	cursor_param := r.URL.Query().Get("cursor")
	if cursor_param != "" {
		var err error
		c.CursorValue, err = strconv.Atoi(cursor_param)
		if err != nil {
			return fmt.Errorf("attempted to parse cursor value %q as int: %w", c.CursorValue, err)
		}
		c.CursorPosition = persistence.CURSOR_MIDDLE
	}
	return nil
}

type FormErrors map[string]string

var formDecoder = form.NewDecoder()
var (
	ErrCorruptedFormData   = errors.New("corrupted form data")
	ErrIncorrectFormParams = errors.New("incorrect form parameters")
)

func parse_form(req *http.Request, result interface{}) error {
	err := req.ParseForm()
	if err != nil {
		return ErrCorruptedFormData
	}

	if err = formDecoder.Decode(result, req.PostForm); err != nil {
		return ErrIncorrectFormParams
	}
	return nil
}
