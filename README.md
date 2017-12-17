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
m.Get("/me/avatar", meAvatarGET)

// or:
m.Get("/admin", adminGET, WithUser, WithAdmin)
```

- Declare routes in var init
```golang
var apiHandler = mux2.NewFromFunc(func(m *mux2.Mux) {
	m.Get("/me", meGET)
})
```


Implementation
--------------
The implementation is simple, but smart: a binary search in a sorted slice to make static routing very fast and simultanuously smallen the search space for dynamic routing. There are few allocations per request, but as we use the normal `http.Handler` interface and need to store parameters on the request's context, there are some. Compared to the very fast [httprouter](https://github.com/julienschmidt/httprouter):
- static routing performs similar
- dynamic routing is not more than 5x slower
- memory usage is around half

TODO
----
- Missing features from `net/http.ServeMux`
  - Host name based matching
  - Multi-goroutine safe
- Add more documentation, publish godoc
- Import benchmarks to this repository
- More tests
- Perhaps: redir "/x" => "/x/" when the former is not defined, but latter is
- Perhaps: `405 method not supported` when handler with other method matches
- Perhaps: think about custom cleanPath, as it is costly
