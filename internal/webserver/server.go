package webserver

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

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

	Profile persistence.Profile
}

func NewApp(profile persistence.Profile) Application {
	ret := Application{
		accessLog: log.New(os.Stdout, "ACCESS\t", log.Ldate|log.Ltime),
		traceLog:  log.New(os.Stdout, "TRACE\t", log.Ldate|log.Ltime),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),

		Profile: profile,
		// formDecoder: form.NewDecoder(),
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

var this_dir string

func init() {
	_, this_file, _, _ := runtime.Caller(0) // `this_file` is absolute path to this source file
	this_dir = path.Dir(this_file)
}

func get_filepath(s string) string {
	return path.Join(this_dir, s)
}

// Manual router implementation.
// I don't like the weird matching behavior of http.ServeMux, and it's not hard to write by hand.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		app.Home(w, r)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[1:]
	switch parts[0] {
	case "static":
		http.StripPrefix("/static", http.FileServer(http.Dir(get_filepath("static")))).ServeHTTP(w, r)
	case "tweet":
		app.TweetDetail(w, r)
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
	err := srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Home' handler (path: %q)", r.URL.Path)
	tpl, err := template.ParseFiles(
		get_filepath("tpl/includes/base.tpl"),
		get_filepath("tpl/home.tpl"),
	)
	panic_if(err)
	app.buffered_render(w, tpl, nil)
}

type TweetDetailData struct {
	persistence.TweetDetailView
	MainTweetID scraper.TweetID
}

func NewTweetDetailData() TweetDetailData {
	return TweetDetailData{
		TweetDetailView: persistence.NewTweetDetailView(),
	}
}
func (t TweetDetailData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t TweetDetailData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}

func to_json(t interface{}) string {
	js, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(js)
}

func (app *Application) TweetDetail(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'TweetDetail' handler (path: %q)", r.URL.Path)
	_, tail := path.Split(r.URL.Path)
	tweet_id, err := strconv.Atoi(tail)
	if err != nil {
		app.error_400_with_message(w, fmt.Sprintf("Invalid tweet ID: %q", tail))
		return
	}

	data := NewTweetDetailData()
	data.MainTweetID = scraper.TweetID(tweet_id)

	trove, err := app.Profile.GetTweetDetail(data.MainTweetID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotInDB) {
			app.error_404(w)
			return
		} else {
			panic(err)
		}
	}
	app.InfoLog.Printf(to_json(trove))
	data.TweetDetailView = trove

	app.buffered_render_template_for(w, "tpl/tweet_detail.tpl", data)
}

type UserProfileData struct {
	persistence.Feed
	scraper.UserID
}

func (t UserProfileData) Tweet(id scraper.TweetID) scraper.Tweet {
	return t.Tweets[id]
}
func (t UserProfileData) User(id scraper.UserID) scraper.User {
	return t.Users[id]
}

func (app *Application) UserFeed(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'UserFeed' handler (path: %q)", r.URL.Path)

	_, tail := path.Split(r.URL.Path)

	user, err := app.Profile.GetUserByHandle(scraper.UserHandle(tail))
	if err != nil {
		app.error_404(w)
		return
	}
	feed, err := app.Profile.GetUserFeed(user.ID, 50, scraper.TimestampFromUnix(0))
	if err != nil {
		panic(err)
	}

	data := UserProfileData{Feed: feed, UserID: user.ID}
	app.InfoLog.Printf(to_json(data))

	app.buffered_render_template_for(w, "tpl/user_feed.tpl", data)
}
