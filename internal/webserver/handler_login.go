package webserver

import (
	"fmt"
	"net/http"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type LoginData struct {
	LoginForm
	ExistingSessions []scraper.UserHandle
}

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

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	app.traceLog.Printf("'Login' handler (path: %q)", r.URL.Path)
	var form LoginForm
	if r.Method == "POST" {
		err := parse_form(r, &form)
		if err != nil {
			app.InfoLog.Print("Form error parse: " + err.Error())
			app.error_400_with_message(w, err.Error())
		}
		form.Validate()
		if len(form.FormErrors) == 0 {
			api := scraper.NewGuestSession()
			api.LogIn(form.Username, form.Password)
			app.Profile.SaveSession(api)
			if err := app.SetActiveUser(api.UserHandle); err != nil {
				app.ErrorLog.Printf(err.Error())
			}
			http.Redirect(w, r, "/login", 303)
		}
		return
	}

	// method = "GET"
	data := LoginData{
		LoginForm:        form,
		ExistingSessions: app.Profile.ListSessions(),
	}
	app.buffered_render_page(w, "tpl/login.tpl", PageGlobalData{}, &data)
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
	err = app.SetActiveUser(scraper.UserHandle(form.AccountName))
	if err != nil {
		app.error_400_with_message(w, fmt.Sprintf("User not in database: %s", form.AccountName))
		return
	}
	app.buffered_render_htmx(w, "nav-sidebar", PageGlobalData{}, nil)
}
