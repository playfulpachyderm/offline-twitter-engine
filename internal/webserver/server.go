package webserver

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
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

	Profile         persistence.Profile
	ActiveUser      scraper.User
	DisableScraping bool
}

func NewApp(profile persistence.Profile) Application {
	ret := Application{
		accessLog: log.New(os.Stdout, "ACCESS\t", log.Ldate|log.Ltime),
		traceLog:  log.New(os.Stdout, "TRACE\t", log.Ldate|log.Ltime),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),

		Profile:    profile,
		ActiveUser: get_default_user(),
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
	case "content":
		http.StripPrefix("/content", http.FileServer(http.Dir(app.Profile.ProfileDir))).ServeHTTP(w, r)
	case "login":
		app.Login(w, r)
	case "change-session":
		app.ChangeSession(w, r)
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
	app.buffered_render_basic_page(w, "tpl/home.tpl", nil)
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
func (t TweetDetailData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}
func (t TweetDetailData) FocusedTweetID() scraper.TweetID {
	return t.MainTweetID
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
	val, err := strconv.Atoi(tail)
	if err != nil {
		app.error_400_with_message(w, fmt.Sprintf("Invalid tweet ID: %q", tail))
		return
	}
	tweet_id := scraper.TweetID(val)

	data := NewTweetDetailData()
	data.MainTweetID = tweet_id

	// Return whether the scrape succeeded (if false, we should 404)
	try_scrape_tweet := func() bool {
		if app.DisableScraping {
			return false
		}
		trove, err := scraper.GetTweetFullAPIV2(tweet_id, 50) // TODO: parameterizable
		if err != nil {
			app.ErrorLog.Print(err)
			return false
		}
		app.Profile.SaveTweetTrove(trove)
		return true
	}

	tweet, err := app.Profile.GetTweetById(tweet_id)
	if err != nil {
		if errors.Is(err, persistence.ErrNotInDB) {
			if !try_scrape_tweet() {
				app.error_404(w)
				return
			}
		} else {
			panic(err)
		}
	} else if !tweet.IsConversationScraped {
		try_scrape_tweet() // If it fails, we can still render it (not 404)
	}

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

	app.buffered_render_tweet_page(w, "tpl/tweet_detail.tpl", data)
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
func (t UserProfileData) Retweet(id scraper.TweetID) scraper.Retweet {
	return t.Retweets[id]
}

func (t UserProfileData) FocusedTweetID() scraper.TweetID {
	return scraper.TweetID(0)
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
		if errors.Is(err, persistence.ErrEndOfFeed) {
			// TODO
		} else {
			panic(err)
		}
	}
	feed.Users[user.ID] = user

	data := UserProfileData{Feed: feed, UserID: user.ID}
	app.InfoLog.Printf(to_json(data))

	app.buffered_render_tweet_page(w, "tpl/user_feed.tpl", data)
}

type FormErrors map[string]string

type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
	FormErrors
}

func (f *LoginForm) Validate() {
	if f.FormErrors == nil {
		f.FormErrors = make(FormErrors)
	}
	if len(f.Username) == 0 {
		f.FormErrors["username"] = "cannot be blank"
	}
	if len(f.Password) == 0 {
		f.FormErrors["password"] = "cannot be blank"
	}
}

type LoginData struct {
	LoginForm
	ExistingSessions []scraper.UserHandle
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Login' handler (path: %q)", r.URL.Path)
	var form LoginForm
	if r.Method == "POST" {
		err := parse_form(r, &form)
		if err != nil {
			app.InfoLog.Print("Form error parse: " + err.Error())
			app.error_400_with_message(w, err.Error())
			return
		}
		form.Validate()
		if len(form.FormErrors) == 0 {
			api := scraper.NewGuestSession()
			api.LogIn(form.Username, form.Password)
			scraper.InitApi(api)
			app.Profile.SaveSession(api)
			http.Redirect(w, r, "/login", 303)
		}
	}
	data := LoginData{
		LoginForm:        form,
		ExistingSessions: app.Profile.ListSessions(),
	}
	app.buffered_render_basic_page(w, "tpl/login.tpl", &data)
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

func (app *Application) ChangeSession(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'change-session' handler (path: %q)", r.URL.Path)
	form := struct {
		AccountName string `form:"account"`
	}{}
	err := parse_form(r, &form)
	if err != nil {
		app.InfoLog.Print("Form error parse: " + err.Error())
		app.error_400_with_message(w, err.Error())
		return
	}
	if form.AccountName == "no account" {
		// Special value that indicates to use a guest session
		scraper.InitApi(scraper.NewGuestSession())
		app.ActiveUser = get_default_user()
		app.DisableScraping = true // API requests will fail b/c not logged in
	} else {
		// Activate the selected session
		user, err := app.Profile.GetUserByHandle(scraper.UserHandle(form.AccountName))
		if err != nil {
			app.error_400_with_message(w, fmt.Sprintf("User not in database: %s", form.AccountName))
			return
		}
		scraper.InitApi(app.Profile.LoadSession(scraper.UserHandle(form.AccountName)))
		app.ActiveUser = user
		app.DisableScraping = false
	}
	app.buffered_render_basic_htmx(w, "nav-sidebar", nil)
}

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
