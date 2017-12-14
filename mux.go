package mux2

import (
	"context"
	"net/http"
	"path"
)

type muxEntry struct {
	method  string
	pattern string
	h       http.Handler
}

type Mux struct {
	m  []muxEntry
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

func (m *Mux) Push(mw Middleware) {
	m.mw = append([]Middleware{mw}, m.mw...)
}

func (m *Mux) Pop(_ Middleware) {
	// parameter is just for esthetics
	m.mw = m.mw[1:]
}

func (m *Mux) handle(method, pattern string, h http.Handler, mw ...Middleware) {
	for _, x := range append(mw, m.mw...) {
		h = x(h)
	}
	m.m = append(m.m, muxEntry{method, pattern, h})
}

func (m *Mux) Handle(p string, h http.Handler, mw ...Middleware) {
	m.handle("", p, h, mw...)
}

func (m *Mux) HandleFunc(p string, h func(http.ResponseWriter, *http.Request), mw ...Middleware) {
	m.handle("", p, http.HandlerFunc(h), mw...)
}

func (m *Mux) Get(p string, h http.Handler, mw ...Middleware)    { m.handle("GET", p, h, mw...) }
func (m *Mux) Post(p string, h http.Handler, mw ...Middleware)   { m.handle("POST", p, h, mw...) }
func (m *Mux) Put(p string, h http.Handler, mw ...Middleware)    { m.handle("PUT", p, h, mw...) }
func (m *Mux) Delete(p string, h http.Handler, mw ...Middleware) { m.handle("DELETE", p, h, mw...) }

type paramsKey struct{}

func (m Mux) handler(host, method, path string) (http.Handler, string, map[string]string) {
	var b *muxEntry
	var p map[string]string
	for i, e := range m.m {
		if e.method != "" && e.method != method {
			continue
		}
		if b != nil && b.pattern > e.pattern {
			// priority is established by comparing the pattern strings, this works because
			// "abc" > "ab"
			// "a/b/c" > "a/:wildcard/c"
			continue
		}
		if ok, p0 := match(e.pattern, path); ok {
			b = &m.m[i]
			p = p0
		}
	}
	if b == nil {
		return http.NotFoundHandler(), "", nil
	}
	h := b.h
	return h, b.pattern, p
}

func match(pat, str string) (bool, map[string]string) {
	var p, s int
	pr := map[string]string{}
	for {
		if p == len(pat) && s == len(str) {
			// precise match
			return true, pr
		} else if p == len(pat) && p > 0 && pat[p-1] == '/' {
			// pattern ending with /, remaining string
			return true, pr
		} else if p == len(pat) || s == len(str) {
			// running out of pattern or string
			return false, nil
		} else if pat[p] == ':' {
			p0 := p
			s0 := s
			for p != len(pat) && pat[p] != '/' {
				p++
			}
			for s != len(str) && str[s] != '/' {
				s++
			}
			pr[pat[p0+1:p]] = str[s0:s]
		} else if pat[p] != str[s] {
			return false, nil
		} else {
			s++
			p++
		}
	}
}

func (m Mux) Handler(r *http.Request) (http.Handler, string, map[string]string) {
	path := cleanPath(r.URL.Path)
	if path != r.URL.Path {
		_, pattern, _ := m.handler(r.Host, r.Method, path)
		url := *r.URL
		url.Path = path
		return http.RedirectHandler(url.String(), http.StatusMovedPermanently), pattern, nil
	}
	return m.handler(r.Host, r.Method, r.URL.Path)
}

func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		// following net/http implementation
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, _, p := m.Handler(r)
	if p != nil {
		r = r.WithContext(context.WithValue(r.Context(), paramsKey{}, p))
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
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

func Params(r *http.Request) map[string]string {
	p, _ := r.Context().Value(paramsKey{}).(map[string]string)
	return p
}

func Param(r *http.Request, key string) string {
	if p := Params(r); p != nil {
		return p[key]
	}
	return ""
}
