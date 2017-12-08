package mux2

import (
	"context"
	"net/http"
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

func (m *Mux) Get(p string, h http.Handler, mw ...Middleware)    { m.handle("GET", p, h, mw...) }
func (m *Mux) Post(p string, h http.Handler, mw ...Middleware)   { m.handle("POST", p, h, mw...) }
func (m *Mux) Put(p string, h http.Handler, mw ...Middleware)    { m.handle("PUT", p, h, mw...) }
func (m *Mux) Delete(p string, h http.Handler, mw ...Middleware) { m.handle("DELETE", p, h, mw...) }

type paramKey struct{}

func (m Mux) Handler(r *http.Request) (http.Handler, string) {
	var b *muxEntry
	var p map[string]string
	for i, e := range m.m {
		if e.method != "" && e.method != r.Method {
			continue
		}
		if b != nil && b.pattern > e.pattern {
			// priority is established by comparing the pattern strings, this works because
			// "abc" > "ab"
			// "a/b/c" > "a/:wildcard/c"
			continue
		}
		if ok, p0 := match(e.pattern, r.URL.Path); ok {
			b = &m.m[i]
			p = p0
		}
	}
	if b == nil {
		return http.NotFoundHandler(), ""
	}
	h := b.h
	if p != nil {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), paramKey{}, p))
			b.h.ServeHTTP(w, r)
		})
	}
	return h, b.pattern
}

func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, _ := m.Handler(r)
	h.ServeHTTP(w, r)
}

func Param(r *http.Request, key string) string {
	if p, _ := r.Context().Value(paramKey{}).(map[string]string); p != nil {
		return p[key]
	}
	return ""
}
