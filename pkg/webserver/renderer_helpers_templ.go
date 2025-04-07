// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package webserver

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig/v3"
	"html/template"
	"net/http"
	"net/http/httptest"

	diffs "gitlab.com/offline-twitter/twitter_offline_engine/pkg/webserver/tmp_templ_diff"
)

// Render the "base" template, creating a full HTML page corresponding to the given template file,
// with all available partials.
func (app *Application) buffered_render_page2(w http.ResponseWriter, tpl_file string, global_data PageGlobalData, tpl_data interface{}) {
	partials := append(glob("tpl/includes/*.tpl"), glob("tpl/tweet_page_includes/*.tpl")...)

	global_data.NotificationBubbles.NumMessageNotifications = len(app.Profile.GetUnreadConversations(app.ActiveUser.ID))
	if app.LastReadNotificationSortIndex != 0 {
		global_data.NotificationBubbles.NumRegularNotifications = app.Profile.GetUnreadNotificationsCount(
			app.ActiveUser.ID,
			app.LastReadNotificationSortIndex,
		)
	}
	global_data.ActiveUser = app.ActiveUser

	filenames := append(partials, get_filepath(tpl_file))

	var tpl *template.Template
	var err error

	_funcs := app.make_funcmap(global_data)
	funcs := sprig.FuncMap()
	for i := range _funcs {
		funcs[i] = _funcs[i]
	}
	if use_embedded == "true" {
		tpl, err = template.New("").Funcs(funcs).ParseFS(embedded_files, filenames...)
	} else {
		tpl, err = template.New("").Funcs(funcs).ParseFiles(filenames...)
	}
	panic_if(err)

	main_component := templ.FromGoHTML(tpl.Lookup("main"), tpl_data)
	b := bytes.Buffer{}
	tpl.Lookup("title").Execute(&b, tpl_data)
	b.WriteString(" | Offline Twitter")
	global_data.Title = b.String()

	buf := new(bytes.Buffer)
	Base(tpl, global_data, main_component).Render(context.Background(), buf)

	// Check against old version
	rec := httptest.NewRecorder()
	app.buffered_render_page(rec, tpl_file, global_data, tpl_data)
	diff, err := diffs.DiffStrings(rec.Body.String(), buf.String())
	panic_if(err)
	if diff != "" {
		println(diff)
		panic(diff)
	}

	_, err = buf.WriteTo(w)
	panic_if(err)
}

func Base(go_tpl *template.Template, global_data PageGlobalData, main_component templ.Component) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(global_data.Title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `pkg/webserver/renderer_helpers.templ`, Line: 73, Col: 29}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</title><link rel=\"stylesheet\" href=\"/static/styles.css\"><link rel=\"shortcut icon\" href=\"/static/twitter.ico\" type=\"image/x-icon\"><link rel=\"stylesheet\" href=\"/static/vendor/fonts.css\"><link rel=\"manifest\" href=\"/static/pwa/manifest.json\"><script src=\"/static/vendor/htmx.min.js\" integrity=\"sha384-zUfuhFKKZCbHTY6aRR46gxiqszMk5tcHjsVFxnUo8VMus4kHGVdIYVbOYYNlKmHV\" crossorigin=\"anonymous\"></script><script src=\"/static/vendor/htmx-extension-json-enc.js\"></script><script>\n\t\t\t\tif ('serviceWorker' in navigator) {\n\t\t\t\t\tnavigator.serviceWorker.register('/static/pwa/service-worker.js')\n\t\t\t\t\t\t.then(function(registration) {\n\t\t\t\t\t\t\tconsole.log('Service Worker registered with scope:', registration.scope);\n\t\t\t\t\t\t}).catch(function(error) {\n\t\t\t\t\t\t\tconsole.log('Service Worker registration failed:', error);\n\t\t\t\t\t\t});\n\t\t\t\t}\n\t\t\t</script><script>\n\t\t\t\t// Set default scrolling (\"instant\", \"smooth\" or \"auto\")\n\t\t\t\thtmx.config.scrollBehavior = \"instant\";\n\n\t\t\t\tdocument.addEventListener('DOMContentLoaded', function() {\n\t\t\t\t\t/**\n\t\t\t\t\t * Consider HTTP 4xx and 500 errors to contain valid HTMX, and swap them as usual\n\t\t\t\t\t */\n\t\t\t\t\tdocument.body.addEventListener('htmx:beforeSwap', function(e) {\n\t\t\t\t\t\tif (e.detail.xhr.status === 500) {\n\t\t\t\t\t\t\te.detail.shouldSwap = true;\n\t\t\t\t\t\t\te.detail.isError = true;\n\t\t\t\t\t\t} else if (e.detail.xhr.status >= 400 && e.detail.xhr.status < 500) {\n\t\t\t\t\t\t\te.detail.shouldSwap = true;\n\t\t\t\t\t\t\te.detail.isError = false;\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t\t});\n\t\t\t</script></head><body><header class=\"row search-bar\"><a onclick=\"window.history.back()\" class=\"button search-bar__back-button\"><img class=\"svg-icon\" src=\"/static/icons/back.svg\" width=\"24\" height=\"24\"></a><form class=\"search-bar__form\" hx-get=\"/search\" hx-push-url=\"true\" hx-target=\"body\" hx-swap=\"innerHTML show:window:top\"><input id=\"searchBar\" class=\"search-bar__input\" name=\"q\" placeholder=\"Search\" type=\"text\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if global_data.GetSearchText() != "" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(global_data.GetSearchText())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `pkg/webserver/renderer_helpers.templ`, Line: 122, Col: 41}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" required></form></header>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = NavSidebar(go_tpl, global_data).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<main>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = main_component.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main><dialog id=\"image_carousel\" class=\"image-carousel\" onmousedown=\"event.button == 0 &amp;&amp; event.target==this &amp;&amp; this.close()\"><div class=\"image-carousel__padding\"><a class=\"button image-carousel__close-button\" onclick=\"image_carousel.close()\">X</a> <img class=\"image-carousel__active-image\" src=\"\"></div></dialog><div class=\"toasts\" id=\"toasts\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, toast := range global_data.Toasts {
			templ_7745c5c3_Err = ToastTpl(toast).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
