package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/Masterminds/sprig/v3"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func panic_if(err error) {
	if err != nil {
		panic(err)
	}
}

var this_dir string

func init() {
	_, this_file, _, _ := runtime.Caller(0) // `this_file` is absolute path to this source file
	this_dir = path.Dir(this_file)
}

func get_filepath(s string) string {
	if use_embedded == "true" {
		return s
	}
	return path.Join(this_dir, s)
}

func glob(path string) []string {
	var ret []string
	var err error
	if use_embedded == "true" {
		ret, err = fs.Glob(embedded_files, get_filepath(path))
	} else {
		ret, err = filepath.Glob(get_filepath(path))
	}
	panic_if(err)
	return ret
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

type TweetCollection interface {
	Tweet(id scraper.TweetID) scraper.Tweet
	User(id scraper.UserID) scraper.User
	Retweet(id scraper.TweetID) scraper.Retweet
	Space(id scraper.SpaceID) scraper.Space
	FocusedTweetID() scraper.TweetID
}

// Creates a template from the given template file using all the available partials.
// Calls `app.buffered_render` to render the created template.
func (app *Application) buffered_render_tweet_page(w http.ResponseWriter, tpl_file string, data TweetCollection) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	r := renderer{
		Funcs: func_map(template.FuncMap{
			"tweet":            data.Tweet,
			"user":             data.User,
			"retweet":          data.Retweet,
			"space":            data.Space,
			"active_user":      app.get_active_user,
			"focused_tweet_id": data.FocusedTweetID,
		}),
		Filenames: append(partials, get_filepath(tpl_file)),
		TplName:   "base",
		Data:      data,
	}
	r.BufferedRender(w)
}

// Creates a template from the given template file using all the available partials.
// Calls `app.buffered_render` to render the created template.
func (app *Application) buffered_render_basic_page(w http.ResponseWriter, tpl_file string, data interface{}) {
	partials := glob("tpl/includes/*.tpl")

	r := renderer{
		Funcs:     func_map(template.FuncMap{"active_user": app.get_active_user}),
		Filenames: append(partials, get_filepath(tpl_file)),
		TplName:   "base",
		Data:      data,
	}
	r.BufferedRender(w)
}

func (app *Application) buffered_render_tweet_htmx(w http.ResponseWriter, tpl_name string, data TweetCollection) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	r := renderer{
		Funcs: func_map(template.FuncMap{
			"tweet":            data.Tweet,
			"user":             data.User,
			"retweet":          data.Retweet,
			"space":            data.Space,
			"active_user":      app.get_active_user,
			"focused_tweet_id": data.FocusedTweetID,
		}),
		Filenames: partials,
		TplName:   tpl_name,
		Data:      data,
	}
	r.BufferedRender(w)
}

func (app *Application) buffered_render_basic_htmx(w http.ResponseWriter, tpl_name string, data interface{}) {
	partials := glob("tpl/includes/*.tpl")

	r := renderer{
		Funcs:     func_map(template.FuncMap{"active_user": app.get_active_user}),
		Filenames: partials,
		TplName:   tpl_name,
		Data:      data,
	}
	r.BufferedRender(w)
}

func (app *Application) get_active_user() scraper.User {
	return app.ActiveUser
}

func func_map(extras template.FuncMap) template.FuncMap {
	ret := sprig.FuncMap()
	for i := range extras {
		ret[i] = extras[i]
	}
	return ret
}

type renderer struct {
	Funcs     template.FuncMap
	Filenames []string
	TplName   string
	Data      interface{}
}

// Render the given template using a bytes.Buffer.  This avoids the possibility of failing partway
// through the rendering, and sending an imcomplete response with "Bad Request" or "Server Error" at the end.
func (r renderer) BufferedRender(w io.Writer) {
	var tpl *template.Template
	var err error
	if use_embedded == "true" {
		tpl, err = template.New("").Funcs(r.Funcs).ParseFS(embedded_files, r.Filenames...)
	} else {
		tpl, err = template.New("").Funcs(r.Funcs).ParseFiles(r.Filenames...)
	}
	panic_if(err)

	buf := new(bytes.Buffer)
	err = tpl.ExecuteTemplate(buf, r.TplName, r.Data)
	panic_if(err)

	_, err = buf.WriteTo(w)
	panic_if(err)
}
