package mux2

import (
	"context"
	"net/http"
	"path"
	"strings"
)

type entry struct {
	method  string
	pattern string
	h       http.Handler
}

type Mux struct {
	m  []entry
	mw []Middleware
}

type Middleware func(http.Handler) http.Handler

func New() *Mux {
	return &Mux{}
}

func NewFromFunc(f func(*Mux)) http.Handler {
	m := New()
	f(m)
	return m
}

// Push middleware on the middleware stack. Handlers registered hereafter will be
// wrapped by the middleware.
func (m *Mux) Push(mw Middleware) {
	m.mw = append([]Middleware{mw}, m.mw...)
}

// Pop top of the middleware stack. Parameter is just for esthetics.
func (m *Mux) Pop(_ Middleware) {
	m.mw = m.mw[1:]
}

func (m *Mux) HandleMethod(method, pattern string, h http.Handler, mw ...Middleware) {
	// "!" is used as wildcard indicator that compares less than all other chars
	pattern = strings.Replace(pattern, ":", "!", -1)
	for _, x := range append(mw, m.mw...) {
		h = x(h)
	}
	// insert new entry into sorted m.m:
	i := 0
	for i < len(m.m) && m.m[i].pattern > pattern {
		i++
	}
	m.m = append(m.m[:i], append([]entry{{method, pattern, h}}, m.m[i:]...)...)

	/* alternative using sort.Slice:
	m.m = append(m.m, muxEntry{method, pattern, h})
	sort.Slice(m.m, func(i, j int) bool {
		return m.m[i].pattern > m.m[j].pattern
	})*/
}

// Handle registers the handler for the given pattern.
// Any method will be matched.
func (m *Mux) Handle(p string, h http.Handler, mw ...Middleware) {
	m.HandleMethod("", p, h, mw...)
}

func (m *Mux) Get(p string, h http.Handler, mw ...Middleware)   { m.HandleMethod("GET", p, h, mw...) }
func (m *Mux) Post(p string, h http.Handler, mw ...Middleware)  { m.HandleMethod("POST", p, h, mw...) }
func (m *Mux) Put(p string, h http.Handler, mw ...Middleware)   { m.HandleMethod("PUT", p, h, mw...) }
func (m *Mux) Patch(p string, h http.Handler, mw ...Middleware) { m.HandleMethod("PATCH", p, h, mw...) }
func (m *Mux) Delete(p string, h http.Handler, mw ...Middleware) {
	m.HandleMethod("DELETE", p, h, mw...)
}

func (m Mux) handler(host, method, path string) (http.Handler, string) {

	// m.m is reverse-sorted by pattern: "/x/a", "/x/", "/x", "/blob", "/!userID/x", "/"
	// binary search lower bound in m.m
	i, j := 0, len(m.m)
	for i < j {
		h := (i + j) / 2
		if m.m[h].pattern > path {
			i = h + 1
		} else {
			j = h
		}
	}

	// considering path "/123" with the above example, i will now point to /!userID/x
	// (since "/!userID" is the first value less than "/123")

	if i != len(m.m) {
		e := m.m[i]
		if (e.method == "" || e.method == method) && e.pattern == path {
			// fast-path for static routes, since match("abc", /"abc") is a lot slower than "/abc" == "/abc"
			return e.h, e.pattern
		}
	}

	// pattern-match till first match
	for i < len(m.m) {
		e := m.m[i]
		if (e.method == "" || e.method == method) && match(e.pattern, path) {
			return e.h, e.pattern
		}
		i++
	}

	return http.NotFoundHandler(), ""
}

func match(pat, str string) bool {
	var p, s int
	for {
		if p == len(pat) && s == len(str) {
			// precise match
			return true
		} else if p == len(pat) && p > 0 && pat[p-1] == '/' {
			// pattern ending with /, remaining string
			return true
		} else if p == len(pat) || s == len(str) {
			// running out of pattern or string
			return false
		} else if pat[p] == '!' {
			p++
			for p != len(pat) && ((pat[p] >= 'a' && pat[p] <= 'z') || (pat[p] >= 'A' && pat[p] <= 'Z')) {
				p++
			}
			var br byte = '/'
			if p != len(pat) {
				br = pat[p]
			}
			for s != len(str) && str[s] != br {
				s++
			}
		} else if pat[p] != str[s] {
			return false
		} else {
			s++
			p++
		}
	}
}

// Handler returns the handler to use for the given request,
// consulting r.Method, r.Host, and r.URL.Path. It always returns
// a non-nil handler.
func (m Mux) Handler(r *http.Request) (http.Handler, string) {
	path := cleanPath(r.URL.Path)
	if path != r.URL.Path {
		_, pattern := m.handler(r.Host, r.Method, path)
		url := *r.URL
		url.Path = path
		return http.RedirectHandler(url.String(), http.StatusMovedPermanently), pattern
	}
	return m.handler(r.Host, r.Method, r.URL.Path)
}

type paramsCtxKey struct{}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		// following net/http implementation
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, p := m.Handler(r)
	if strings.IndexByte(p, '!') >= 0 {
		r = r.WithContext(context.WithValue(r.Context(), paramsCtxKey{}, params{p, r.URL.Path}))
	}
	h.ServeHTTP(w, r)
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	if strings.IndexByte(p, '.') >= 0 {
		// check because path.Clean is expensive
		np := path.Clean(p)
		// path.Clean removes trailing slash except for root;
		// put the trailing slash back if necessary.
		if p[len(p)-1] == '/' && np != "/" {
			np += "/"
		}
		p = np
	}
	return p
}

type params struct {
	pat, str string
}

func (pp params) Get(key string) string {
	// Meh, this kind-of copies the match() implementation. Ideas welcome.
	var p, s int
	for p != len(pp.pat) && s != len(pp.str) {
		if pp.pat[p] == '!' {

			p++

			p0 := p
			s0 := s

			for p != len(pp.pat) && ((pp.pat[p] >= 'a' && pp.pat[p] <= 'z') || (pp.pat[p] >= 'A' && pp.pat[p] <= 'Z')) {
				p++
			}
			var br byte = '/'
			if p != len(pp.pat) {
				br = pp.pat[p]
			}
			for s != len(pp.str) && pp.str[s] != br {
				s++
			}
			if pp.pat[p0:p] == key {
				return pp.str[s0:s]
			}
		} else {
			s++
			p++
		}
	}
	return ""
}

func Param(r *http.Request, key string) string {
	p, _ := r.Context().Value(paramsCtxKey{}).(params)
	return p.Get(key)
}
