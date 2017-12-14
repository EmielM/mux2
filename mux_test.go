package mux2

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var muxTests = []struct {
	method  string
	path    string
	code    int
	body    string
	pattern string
}{
	{"GET", "/x", 200, "/x", "/x"},
	{"GET", "/./x", 301, "dontCare", "/x"},
	{"PUT", "/x", 404, "dontCare", ""},
	{"GET", "/x/123/abc", 200, "/x/*", "/x/"},
	{"PUT", "/x/12345", 200, "/x/*", "/x/"},
	{"PUT", "/abc/../blob", 301, "dontCare", "/blob"},
	{"PUT", "/blob", 200, "put /blob", "/blob"},
	{"GET", "/x/a", 200, "/x/a", "/x/a"},
	{"GET", "/123/x", 200, "/venue=123/x", "/:venueID/x"},
}

func TestMux(t *testing.T) {

	m := New()
	m.Get("/:venueID/x", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/venue=" + Param(r, "venueID") + "/x"))
	}))
	m.Get("/x/a", stringHandler("/x/a"))
	m.Handle("/x/", stringHandler("/x/*"))
	m.Get("/x", stringHandler("/x"))
	m.Get("/", stringHandler("/"))
	m.Put("/blob", stringHandler("put /blob"))

	for _, tt := range muxTests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(tt.method, tt.path, nil)
		m.ServeHTTP(w, r)
		_, pattern, _ := m.Handler(r)
		if w.Code != tt.code || (tt.body != "dontCare" && w.Body.String() != tt.body) || pattern != tt.pattern {
			t.Errorf("%s %s = %d %s (%s), want %d %s (%s)", tt.method, tt.path, w.Code, w.Body.String(), pattern, tt.code, tt.body, tt.pattern)
		}
	}
}

func stringHandler(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	})
}
