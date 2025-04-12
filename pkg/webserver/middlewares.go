package webserver

import (
	"fmt"
	"net/http"
	"context"
	"time"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/tracing"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Content-Security-Policy",
		//	"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com; img-src 'self' pbs.twimg.com")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		// w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		var span *tracing.Span
		var ctx context.Context
		if is_htmx(r) {
			ctx, span = tracing.InitTrace(r.Context(), "htmx")
		} else {
			ctx, span = tracing.InitTrace(r.Context(), "main")
		}
		r = r.WithContext(ctx)
		defer func() {
			span.End()
		}()

		next.ServeHTTP(w, r)
		duration := time.Since(t)

		app.AccessLog.Printf("%s - %s %s %s\t%s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), duration)
	})
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.error_500(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
