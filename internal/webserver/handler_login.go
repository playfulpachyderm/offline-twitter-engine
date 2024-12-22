package webserver

import (
	"encoding/json"
	"errors"
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
			api, err := scraper.NewGuestSession()
			if err != nil {
				panic(err.Error()) // Return it as a toast
			}
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
	user, err := api.GetUser(api.UserHandle)
	if err != nil { // ErrDoesntExist or otherwise
		app.error_404(w, r)
		return
	}
	panic_if(app.Profile.SaveUser(&user)) // TODO: handle conflicting users
	panic_if(app.Profile.DownloadUserContentFor(&user, &app.API))

	// Now that the user is scraped for sure, set them as the logged-in user
	err = app.SetActiveUser(api.UserHandle)
	panic_if(err)

	// Scrape the user's feed
	trove, err := app.API.GetHomeTimeline("", true)
	if err != nil {
		app.ErrorLog.Printf("Initial timeline scrape failed: %s", err.Error())
		http.Redirect(w, r, "/", 303)
	}
	fmt.Println("Saving initial feed results...")
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)

	// Scrape the user's followers
	trove, err = app.API.GetFollowees(user.ID, 1000)
	if err != nil {
		app.ErrorLog.Printf("Failed to scrape followers: %s", err.Error())
		http.Redirect(w, r, "/", 303)
	}
	app.Profile.SaveTweetTrove(trove, false, &app.API)
	app.Profile.SaveAsFolloweesList(user.ID, trove)
	go app.Profile.SaveTweetTrove(trove, true, &app.API)

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
		app.error_400_with_message(w, r, fmt.Sprintf("User not in database: %s", form.AccountName))
		return
	}
	app.LastReadNotificationSortIndex = 0 // Clear unread notifications

	// Update notifications info in background (avoid latency when switching users)
	go func() {
		trove, last_unread_notification_sort_index, err := app.API.GetNotifications(1) // Just 1 page
		if err != nil && !errors.Is(err, scraper.END_OF_FEED) && !errors.Is(err, scraper.ErrRateLimited) {
			app.ErrorLog.Printf("Error occurred on getting notifications after switching users: %s", err.Error())
			return
		}
		// We have to save the notifications first, otherwise it'll just report 0 since the last-read sort index
		app.Profile.SaveTweetTrove(trove, false, &app.API)
		go app.Profile.SaveTweetTrove(trove, true, &app.API)
		// Set the notifications count
		app.LastReadNotificationSortIndex = last_unread_notification_sort_index
	}()
	data := NotificationBubbles{
		NumMessageNotifications: len(app.Profile.GetUnreadConversations(app.ActiveUser.ID)),
	}
	if app.LastReadNotificationSortIndex != 0 {
		data.NumRegularNotifications = app.Profile.GetUnreadNotificationsCount(app.ActiveUser.ID, app.LastReadNotificationSortIndex)
	}
	app.buffered_render_htmx(w, "nav-sidebar", PageGlobalData{}, data)
}
