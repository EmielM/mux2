[![Build Status](https://travis-ci.org/EmielM/mux2.svg?branch=master)](http://travis-ci.org/EmielM/mux2)

mux2
====

A http router (mux) with the simplicity of `net/http`'s `ServeMux`, but adds:

- Parameters
```golang
m.Handle("/:userID/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userID := mux2.Param(r, "userID")
}))
```

- HTTP method matching:
```golang
m.Handle("/allMethods", h)
m.Put("/path", h)
m.Get("/path", h)
```

- Stacked middleware
```golang
m.Get("/public", publicGET)
m.Push(WithUser)
m.Get("/me", meGET)

// or:
m.Get("/admin", adminGET, WithUser, WithAdmin)
```

- Declare routes in var init
```golang
var apiHandler = mux2.NewFromFunc(func(m *mux2.Mux) {
	m.Get("/me", meGET)
})
```

The implementation features simplicity as well: a linear lookup through all the patterns. Your cpu time is going to be dominated by other stuff anyway.

No radix trees, or constructing regexps. Just 120 lines of code, no dependencies.

TODO
----
- Missing features from `net/http.ServeMux`
  - Host name based matching
  - Multi-goroutine safe
- Perhaps: redir "/x" => "/x/" when the former is not defined, but latter is
- Perhaps: `405 method not supported` when handler with other method matches
- More tests
- Get the word out
