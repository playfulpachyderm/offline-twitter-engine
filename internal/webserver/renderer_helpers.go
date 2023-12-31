package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/Masterminds/sprig/v3"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

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

// Config object for buffered rendering
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

	funcs := sprig.FuncMap()
	for i := range r.Funcs {
		funcs[i] = r.Funcs[i]
	}
	if use_embedded == "true" {
		tpl, err = template.New("").Funcs(funcs).ParseFS(embedded_files, r.Filenames...)
	} else {
		tpl, err = template.New("").Funcs(funcs).ParseFiles(r.Filenames...)
	}
	panic_if(err)

	buf := new(bytes.Buffer)
	err = tpl.ExecuteTemplate(buf, r.TplName, r.Data)
	panic_if(err)

	_, err = buf.WriteTo(w)
	panic_if(err)
}

// Render the "base" template, creating a full HTML page corresponding to the given template file,
// with all available partials.
func (app *Application) buffered_render_page(w http.ResponseWriter, tpl_file string, global_data PageGlobalData, tpl_data interface{}) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	r := renderer{
		Funcs: template.FuncMap{
			"tweet":                  global_data.Tweet,
			"user":                   global_data.User,
			"retweet":                global_data.Retweet,
			"space":                  global_data.Space,
			"active_user":            app.get_active_user,
			"focused_tweet_id":       global_data.GetFocusedTweetID,
			"search_text":            global_data.GetSearchText,
			"get_entities":           get_entities,
			"get_tombstone_text":     get_tombstone_text,
			"cursor_to_query_params": cursor_to_query_params,
		},
		Filenames: append(partials, get_filepath(tpl_file)),
		TplName:   "base",
		Data:      tpl_data,
	}
	r.BufferedRender(w)
}

// Render a particular template (HTMX response, i.e., not a full page)
func (app *Application) buffered_render_htmx(w http.ResponseWriter, tpl_name string, global_data PageGlobalData, tpl_data interface{}) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	r := renderer{
		Funcs: template.FuncMap{
			"tweet":                  global_data.Tweet,
			"user":                   global_data.User,
			"retweet":                global_data.Retweet,
			"space":                  global_data.Space,
			"active_user":            app.get_active_user,
			"focused_tweet_id":       global_data.GetFocusedTweetID,
			"search_text":            global_data.GetSearchText,
			"get_entities":           get_entities,
			"get_tombstone_text":     get_tombstone_text,
			"cursor_to_query_params": cursor_to_query_params,
		},
		Filenames: partials,
		TplName:   tpl_name,
		Data:      tpl_data,
	}
	r.BufferedRender(w)
}

func (app *Application) get_active_user() scraper.User {
	return app.ActiveUser
}

func cursor_to_query_params(c persistence.Cursor) string {
	result := url.Values{}
	result.Set("cursor", fmt.Sprint(c.CursorValue))
	result.Set("sort-order", c.SortOrder.String())
	return result.Encode()
}

type EntityType int

const (
	ENTITY_TYPE_TEXT EntityType = iota
	ENTITY_TYPE_MENTION
	ENTITY_TYPE_HASHTAG
)

type Entity struct {
	EntityType
	Contents string
}

func get_entities(text string) []Entity {
	ret := []Entity{}
	start := 0
	for _, idxs := range regexp.MustCompile(`(\W|^)[@#]\w+`).FindAllStringIndex(text, -1) {
		// Handle leading whitespace.  Only match start-of-string or leading whitespace to avoid matching, e.g., emails
		if text[idxs[0]] == ' ' || text[idxs[0]] == '\n' {
			idxs[0] += 1
		}
		if start != idxs[0] {
			ret = append(ret, Entity{ENTITY_TYPE_TEXT, text[start:idxs[0]]})
		}
		piece := text[idxs[0]+1 : idxs[1]] // Chop off the "#" or "@"
		if text[idxs[0]] == '@' {
			ret = append(ret, Entity{ENTITY_TYPE_MENTION, piece})
		} else {
			ret = append(ret, Entity{ENTITY_TYPE_HASHTAG, piece})
		}
		start = idxs[1]
	}
	if start < len(text) {
		ret = append(ret, Entity{ENTITY_TYPE_TEXT, text[start:]})
	}

	return ret
}

func get_tombstone_text(t scraper.Tweet) string {
	if t.TombstoneText != "" {
		return t.TombstoneText
	}
	return t.TombstoneType
}

// TODO: this name sucks
type PageGlobalData struct {
	scraper.TweetTrove
	SearchText     string
	FocusedTweetID scraper.TweetID
}

func (d PageGlobalData) Tweet(id scraper.TweetID) scraper.Tweet {
	return d.Tweets[id]
}
func (d PageGlobalData) User(id scraper.UserID) scraper.User {
	return d.Users[id]
}
func (d PageGlobalData) Retweet(id scraper.TweetID) scraper.Retweet {
	return d.Retweets[id]
}
func (d PageGlobalData) Space(id scraper.SpaceID) scraper.Space {
	return d.Spaces[id]
}
func (d PageGlobalData) GetFocusedTweetID() scraper.TweetID {
	return d.FocusedTweetID
}
func (d PageGlobalData) GetSearchText() string {
	fmt.Println(d.SearchText)
	return d.SearchText
}
