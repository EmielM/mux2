package mux2

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMux(t *testing.T) {

	m := New()
	m.Get("/:venueID/x", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/venue=" + Param(r, "venueID") + "/x"))
	}))
	m.Get("/x/a", serveString("/x/a"))
	m.Handle("/x/", serveString("/x/*"))
	m.Get("/x", serveString("/x"))
	m.Get("/", serveString("/"))

	doRequest(t, m, "GET", "/x", "/x")
	doRequest(t, m, "PUT", "/x", "404 page not found\n")
	doRequest(t, m, "GET", "/x/123/abc", "/x/*")
	doRequest(t, m, "PUT", "/x/12345", "/x/*")
	doRequest(t, m, "GET", "/x/a", "/x/a")
	doRequest(t, m, "GET", "/123/x", "/venue=123/x")
}

func serveString(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	})
}

func doRequest(t *testing.T, h http.Handler, method, path, expected string) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	b := w.Body.String()
	if b != expected {
		t.Errorf("%s %s got %s, expected %s", method, path, b, expected)
	}
}
