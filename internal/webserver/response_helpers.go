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

func (app *Application) error_500(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	err2 := app.ErrorLog.Output(2, trace) // Magic
	if err2 != nil {
		panic(err2)
	}
	http.Error(w, "Server error :(", 500)
}
