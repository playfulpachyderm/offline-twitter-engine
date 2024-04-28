package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type LoginData struct {
	LoginForm
	ExistingSessions []scraper.UserHandle
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FormErrors
}
type FormErrors map[string]string

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

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Login' handler (path: %q)", r.URL.Path)
	var form LoginForm
	if r.Method == "POST" {
		data, err := io.ReadAll(r.Body)
		panic_if(err)
		panic_if(json.Unmarshal(data, &form)) // TODO: HTTP 400 not 500
		form.Validate()
		if len(form.FormErrors) == 0 {
			api := scraper.NewGuestSession()
			challenge := api.LogIn(form.Username, form.Password)
			if challenge != nil {
				panic( // Middleware will trap this panic and return an HTMX error toast
					"Twitter challenged your login.  Make sure your username is your user handle (not email).  " +
						"If you're logging in from a new device, try doing it on the official Twitter site first, then try again here.")
			}
			app.after_login(w, r, api)
		}
		return
	}

	// method == "GET"
	data := LoginData{
		LoginForm:        form,
		ExistingSessions: app.Profile.ListSessions(),
	}
	app.buffered_render_page(w, "tpl/login.tpl", PageGlobalData{}, &data)
}

func (app *Application) after_login(w http.ResponseWriter, r *http.Request, api scraper.API) {
	app.Profile.SaveSession(api)

	// Ensure the user is downloaded
	user, err := scraper.GetUser(api.UserHandle)
	if err != nil {
		app.error_404(w)
		return
	}
	panic_if(app.Profile.SaveUser(&user))
	panic_if(app.Profile.DownloadUserContentFor(&user))

	// Now that the user is scraped for sure, set them as the logged-in user
	err = app.SetActiveUser(api.UserHandle)
	panic_if(err)

	// Scrape the user's feed
	trove, err := scraper.GetHomeTimeline("", true)
	if err != nil {
		app.ErrorLog.Printf("Initial timeline scrape failed: %s", err.Error())
		http.Redirect(w, r, "/", 303)
	}
	fmt.Println("Saving initial feed results...")
	app.Profile.SaveTweetTrove(trove, false)
	go app.Profile.SaveTweetTrove(trove, true)

	// Scrape the user's followers
	trove, err = scraper.GetFollowees(user.ID, 1000)
	if err != nil {
		app.ErrorLog.Printf("Failed to scrape followers: %s", err.Error())
		http.Redirect(w, r, "/", 303)
	}
	app.Profile.SaveTweetTrove(trove, false)
	app.Profile.SaveAsFolloweesList(user.ID, trove)
	go app.Profile.SaveTweetTrove(trove, true)

	// Redirect to Timeline
	http.Redirect(w, r, "/", 303)
}

func (app *Application) ChangeSession(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'change-session' handler (path: %q)", r.URL.Path)
	form := struct {
		AccountName string `json:"account"`
	}{}
	formdata, err := io.ReadAll(r.Body)
	panic_if(err)
	panic_if(json.Unmarshal(formdata, &form)) // TODO: HTTP 400 not 500
	err = app.SetActiveUser(scraper.UserHandle(form.AccountName))
	if err != nil {
		app.error_400_with_message(w, fmt.Sprintf("User not in database: %s", form.AccountName))
		return
	}
	data := Notifications{NumMessageNotifications: len(app.Profile.GetUnreadConversations(app.ActiveUser.ID))}
	app.buffered_render_htmx(w, "nav-sidebar", PageGlobalData{}, data)
}
