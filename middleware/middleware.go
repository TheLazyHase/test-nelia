package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
)

type Middleware func(next http.Handler) http.Handler

type Chainer struct {
	Middlewares []Middleware
}

func (c *Chainer) Chain(handler http.Handler) http.Handler {
	for _, middleware := range c.Middlewares {
		handler = middleware(handler)
	}

	return handler
}

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqDump, _ := httputil.DumpRequest(r, true)
		log.Printf("REQUEST:\n%s", string(reqDump))
		next.ServeHTTP(w, r)
	})
}
