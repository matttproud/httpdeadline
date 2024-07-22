package httpdeadline

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func newServer(t testing.TB, h http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func urlOf(t testing.TB, srv *httptest.Server) *url.URL {
	t.Helper()
	url, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	return url
}

func newClient() *http.Client {
	return &http.Client{
		Transport: new(http.Transport),
	}
}

func newGetRequest(t testing.TB, url *url.URL) *http.Request {
	t.Helper()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

type spyHandler struct {
	Deadline time.Time
	OK       bool
}

func (h *spyHandler) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	h.Deadline, h.OK = req.Context().Deadline()
}

var now = time.Date(2024, 7, 22, 20, 10, 0, 0, time.UTC)

func asTimeFormat(t time.Time) string { return t.Format(http.TimeFormat) }
func asRFC850(t time.Time) string     { return t.Format(time.RFC850) }
func asANSIC(t time.Time) string      { return t.Format(time.ANSIC) }

func TestFromHeader(t *testing.T) {
	for _, test := range []struct {
		Name string

		Header http.Header

		Status   int
		Deadline time.Time
		OK       bool
	}{
		{
			Name:     "none",
			Header:   nil,
			Status:   200,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "valid-timeformat",
			Header: http.Header{
				"X-MTP-Deadline": []string{asTimeFormat(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "valid-rfc850",
			Header: http.Header{
				"X-MTP-Deadline": []string{asRFC850(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "valid-ansic",
			Header: http.Header{
				"X-MTP-Deadline": []string{asANSIC(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "invalid-timeformat",
			Header: http.Header{
				"X-MTP-Deadline": []string{asTimeFormat(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "invalid-rfc850",
			Header: http.Header{
				"X-MTP-Deadline": []string{asRFC850(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "invalid-ansic",
			Header: http.Header{
				"X-MTP-Deadline": []string{asANSIC(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-timeformat",
			Header: http.Header{
				"X-MTP-Deadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-rfc850",
			Header: http.Header{
				"X-MTP-Deadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-ansic",
			Header: http.Header{
				"X-MTP-Deadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			var spy spyHandler
			h := FromHeader("X-MTP-Deadline", &spy)
			srv := newServer(t, h)
			req := newGetRequest(t, urlOf(t, srv))
			req.Header = test.Header
			client := newClient()
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := resp.StatusCode, test.Status; got != want {
				t.Errorf("resp.StatusCode = %v, want %v", got, want)
			}
			if got, want := spy.Deadline, test.Deadline; !got.Equal(want) {
				t.Errorf("spy.Deadline = %v, want %v", got, want)
			}
			if got, want := spy.OK, test.OK; got != want {
				t.Errorf("spy.OK = %v, want %v", got, want)
			}
		})
	}
}

func TestFromQueryParams(t *testing.T) {
	for _, test := range []struct {
		Name string

		Query url.Values

		Status   int
		Deadline time.Time
		OK       bool
	}{
		{
			Name:     "none",
			Query:    nil,
			Status:   200,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "valid-timeformat",
			Query: url.Values{
				"mtpdeadline": []string{asTimeFormat(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "valid-rfc850",
			Query: url.Values{
				"mtpdeadline": []string{asRFC850(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "valid-ansic",
			Query: url.Values{
				"mtpdeadline": []string{asANSIC(now)},
			},
			Status:   200,
			Deadline: now,
			OK:       true,
		},
		{
			Name: "invalid-timeformat",
			Query: url.Values{
				"mtpdeadline": []string{asTimeFormat(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "invalid-rfc850",
			Query: url.Values{
				"mtpdeadline": []string{asRFC850(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "invalid-ansic",
			Query: url.Values{
				"mtpdeadline": []string{asANSIC(now) + "garbage"},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-timeformat",
			Query: url.Values{
				"mtpdeadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-rfc850",
			Query: url.Values{
				"mtpdeadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
		{
			Name: "empty-ansic",
			Query: url.Values{
				"mtpdeadline": []string{""},
			},
			Status:   400,
			Deadline: time.Time{},
			OK:       false,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			var spy spyHandler
			h := FromQueryParams("mtpdeadline", &spy)
			srv := newServer(t, h)
			url := urlOf(t, srv)
			url.RawQuery = test.Query.Encode()
			req := newGetRequest(t, url)
			client := newClient()
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := resp.StatusCode, test.Status; got != want {
				t.Errorf("resp.StatusCode = %v, want %v", got, want)
			}
			if got, want := spy.Deadline, test.Deadline; !got.Equal(want) {
				t.Errorf("spy.Deadline = %v, want %v", got, want)
			}
			if got, want := spy.OK, test.OK; got != want {
				t.Errorf("spy.OK = %v, want %v", got, want)
			}
		})
	}
}
