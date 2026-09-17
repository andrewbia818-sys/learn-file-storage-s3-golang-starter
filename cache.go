package main

import "net/http"

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
//		w.Header().Set("Pragma", "no-cache")
//		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

/* OLD VERSION BELOW
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=3600")
		next.ServeHTTP(w, r)
	})
}
	*/