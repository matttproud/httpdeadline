// Package httpdeadline adds request deadline support to inbound HTTP requests
// using a microscopic http.Handler chaining middleware.
//
// This package is discussed further at
// https://matttproud.com/blog/posts/context-cancellation-and-server-libraries.html.
//
// # Usage
//
// Wrap the outermost [http.Handler] that you will register with the
// [http.ServeMux] with either [FromHeader] or [FromQueryParams].
//
//	var mux http.ServeMux
//	mux.Handle("/teapotz", httpdeadline.FromHeader("X-MTP-Deadline", teapotz))
//
// If the /teapotz path receives a request with a HTTP header of key
// "X-MTP-Deadline" with a [http.ParseTime]-valid value, that value will be used
// as a deadline for the request that the teapotz handler receives.
//
// As of the implementation of this package, [http.TimeFormat], [time.RFC850],
// and [time.ANSIC] formats can be provided as values.  Values that cannot be
// parsed produce [http.StatusBadRequest] results.
//
// The same principles described above with [FromHeader] apply to
// [FromQueryParams].
//
// # Environmental Considerations
//
// Consider where this package is used and whether it is in a public or private
// system and what sends the system requests.  If it is only trusted parties
// sending your system requests, this package is probably safe to use.
//
// Tread carefully with public systems or with untrusted users.  It is possible
// to perform somewhat malicious things using incorrect context deadlines (e.g.,
// exhaust underlying backend systems by allowing them to continue for too
// long).
package httpdeadline

import (
	"context"
	"net/http"
)

// FromHeader wraps the provided [http.Handler] in an outer http.Handler that
// sets a maximum a deadline on the [http.Request]'s context if the named HTTP
// header is set to a [http.ParseTime]-compatible value.  That value becomes the
// maximum deadline for the request.
func FromHeader(name string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := req.Header[http.CanonicalHeaderKey(name)]; !ok {
			h.ServeHTTP(w, req)
			return
		}
		val := req.Header.Get(name)
		time, err := http.ParseTime(val)
		if val == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithDeadline(req.Context(), time)
		defer cancel()
		h.ServeHTTP(w, req.WithContext(ctx))
	})
}

// FromQueryParams wraps the provided [http.Handler] in an outer http.Handler
// that sets a maximum a deadline on the [http.Request]'s context if the named
// query parameter is set to a [http.ParseTime]-compatible value.  That value
// becomes the maximum deadline for the request.
func FromQueryParams(name string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !req.URL.Query().Has(name) {
			h.ServeHTTP(w, req)
			return
		}
		val := req.URL.Query().Get(name)
		time, err := http.ParseTime(val)
		if val == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithDeadline(req.Context(), time)
		defer cancel()
		h.ServeHTTP(w, req.WithContext(ctx))
	})
}
