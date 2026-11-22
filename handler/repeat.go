package handler

import (
	"io"
	"log"
	"net/http"
)

type RepeatHandler struct {
}

// this handler is only for tests
func (_ RepeatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(msg)

	return
}
