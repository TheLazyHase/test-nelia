package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/esequiel378/hl7"
)

type Middleware func(next http.Handler) http.Handler

func Chain(middleware ...Middleware) http.Handler {
	var handler http.Handler
	for i := range middleware {
		handler = middleware[len(middleware)-1-i](handler)
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
