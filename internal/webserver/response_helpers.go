package webserver

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func panic_if(err error) {
	if err != nil {
		panic(err)
	}
}

func (app *Application) error_400_with_message(w http.ResponseWriter, r *http.Request, msg string) {
	if is_htmx(r) {
		app.toast(w, r, 400, Toast{Title: "Bad Request", Message: msg, Type: "error"})
	} else {
		http.Error(w, fmt.Sprintf("Bad Request\n\n%s", msg), 400)
	}
}

func (app *Application) error_401(w http.ResponseWriter, r *http.Request) {
	msg := "Please log in or set an active session."
	if app.ActiveUser.ID != 0 {
		msg += "  (There is currently an active user, but scraping is disabled.)"
	}
	if is_htmx(r) {
		app.toast(w, r, 401, Toast{Title: "Login required", Message: msg, Type: "error"})
	} else {
		http.Error(w, msg, 401)
	}
}

func (app *Application) error_404(w http.ResponseWriter, r *http.Request) {
	if is_htmx(r) {
		app.toast(w, r, 404, Toast{Title: "Not found", Type: "error"})
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func (app *Application) error_500(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	err2 := app.ErrorLog.Output(2, trace) // Magic
	if err2 != nil {
		panic(err2)
	}

	if is_htmx(r) {
		app.toast(w, r, 500, Toast{Title: "Server error", Message: err.Error(), Type: "error"})
	} else {
		http.Error(w, err.Error(), 500)
	}
}

// The toast is the primary payload (i.e., not OOB)
func (app *Application) toast(w http.ResponseWriter, r *http.Request, status_code int, t Toast) {
	// Reset the HTMX response to return an error toast and append it to the Toasts container
	w.Header().Set("HX-Reswap", "beforeend")
	w.Header().Set("HX-Retarget", "#toasts")
	w.Header().Set("HX-Push-Url", "false")

	w.WriteHeader(status_code) // Must be called after all `Header.Set(...)`

	app.buffered_render_htmx(w, "toast", PageGlobalData{}, t)
}

// `Type` can be:
//   - "success" (default)
//   - "warning"
//   - "error"
//
// If "AutoCloseDelay" is not 0, the toast will auto-disappear after that many milliseconds.
type Toast struct {
	Title          string
	Message        string
	Type           string
	AutoCloseDelay int64
}
