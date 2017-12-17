[![Build Status](https://travis-ci.org/EmielM/mux2.svg?branch=master)](http://travis-ci.org/EmielM/mux2)

mux2
====

A http router (mux) with the simplicity of `net/http`'s `ServeMux`, but adds:

- Parameters
```golang
m.Handle("/:userID/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userID := mux2.Param(r, "userID")
}))

// You can use parameters and static routes together:
m.Handle("/user/me", meHandler)
m.Handle("/user/:userID", userHandler)
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

Implementation & performance
----------------------------
The implementation is simple: requests are routed using a sorted slice of mux entries.
- Static routing is fast and simple: a binary search in a contiguous slice.
- Dynamic routing can still use the sorted slice to prune a lot of the search space, especially for common use cases.
- Unintuitively, it uses roughly half the memory as implementations that use radix tries.
- There are no heap allocations per request, except those incurred by request.WithContext(). That call alone incurs a 5x performance hit for dynamic routes, though. I don't see a way to prevent this while sticking to the `http.Handler` interface.

All in all, compared to the very fast [httprouter](https://github.com/julienschmidt/httprouter):
- static routing performs similar
- dynamic routing is roughly 5x slower (httprouter's has an own Handler type)
- memory usage is around half
- approximately 5x less code

(tests based on [go-http-routing-benchmark](https://github.com/emielm/go-http-routing-benchmark).)

TODO
----
- Add documentation in code, publish godoc
- More tests
- Import benchmarks to this repository: is there a way to do it without depending on httprouter?
- Perhaps: redir "/x" => "/x/" when the former is not defined, but latter is
- Perhaps: `405 method not supported` when handler with other method matches
- Perhaps: multi-goroutine safe registration (like `net/http.ServeMux`)
- Perhaps: host name based matching (like `net/http.ServeMux`)
