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

// func (app *Application) error_400(w http.ResponseWriter) {
// 	http.Error(w, "Bad Request", 400)
// }

func (app *Application) error_400_with_message(w http.ResponseWriter, msg string) {
	http.Error(w, fmt.Sprintf("Bad Request\n\n%s", msg), 400)
}

func (app *Application) error_401(w http.ResponseWriter) {
	http.Error(w, "Please log in or set an active session", 401)
}

func (app *Application) error_404(w http.ResponseWriter) {
	http.Error(w, "Not Found", 404)
}

func (app *Application) error_500(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	err2 := app.ErrorLog.Output(2, trace) // Magic
	if err2 != nil {
		panic(err2)
	}
	app.toast(w, r, Toast{Title: "Server error", Message: err.Error(), Type: "error"})
}

func (app *Application) toast(w http.ResponseWriter, r *http.Request, t Toast) {
	// Reset the HTMX response to return an error toast and append it to the Toasts container
	w.Header().Set("HX-Reswap", "beforeend")
	w.Header().Set("HX-Retarget", "#toasts")
	w.Header().Set("HX-Push-Url", "false")

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
