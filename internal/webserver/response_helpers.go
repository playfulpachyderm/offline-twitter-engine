package webserver

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime/debug"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
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

// Render the given template using a bytes.Buffer.  This avoids the possibility of failing partway
// through the rendering, and sending an imcomplete response with "Bad Request" or "Server Error" at the end.
func (app *Application) buffered_render(w http.ResponseWriter, tpl *template.Template, data interface{}) {
	// The template to render is always "base".  The choice of which template files to parse into the
	// template affects what the contents of "main" (inside of "base") will be
	buf := new(bytes.Buffer)
	err := tpl.ExecuteTemplate(buf, "base", data)
	panic_if(err)

	_, err = buf.WriteTo(w)
	panic_if(err)
}

type TweetCollection interface {
	Tweet(id scraper.TweetID) scraper.Tweet
	User(id scraper.UserID) scraper.User
	FocusedTweetID() scraper.TweetID
}

// Creates a template from the given template file using all the available partials.
// Calls `app.buffered_render` to render the created template.
func (app *Application) buffered_render_tweet_page(w http.ResponseWriter, tpl_file string, data TweetCollection) {
	partials, err := filepath.Glob(get_filepath("tpl/includes/*.tpl"))
	panic_if(err)
	tweet_partials, err := filepath.Glob(get_filepath("tpl/tweet_page_includes/*.tpl"))
	panic_if(err)
	partials = append(partials, tweet_partials...)

	tpl, err := template.New("does this matter at all? lol").Funcs(
		template.FuncMap{"tweet": data.Tweet, "user": data.User, "active_user": app.get_active_user, "focused_tweet_id": data.FocusedTweetID},
	).ParseFiles(append(partials, get_filepath(tpl_file))...)
	panic_if(err)

	app.buffered_render(w, tpl, data)
}

// Creates a template from the given template file using all the available partials.
// Calls `app.buffered_render` to render the created template.
func (app *Application) buffered_render_basic_page(w http.ResponseWriter, tpl_file string, data interface{}) {
	partials, err := filepath.Glob(get_filepath("tpl/includes/*.tpl"))
	panic_if(err)

	tpl, err := template.New("does this matter at all? lol").Funcs(
		template.FuncMap{"active_user": app.get_active_user},
	).ParseFiles(append(partials, get_filepath(tpl_file))...)
	panic_if(err)

	app.buffered_render(w, tpl, data)
}

func (app *Application) get_active_user() scraper.User {
	return app.ActiveUser
}
