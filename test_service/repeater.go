package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

// To simulate the lab software, I use a simple repeater, who repeat the content of the request if it's POST and its URL if it's not
func main() {
	//Match everything
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqDump, _ := httputil.DumpRequest(r, true)
		log.Printf("REQUEST:\n%s", string(reqDump))
		if r.Method == "POST" {
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.String()))
		return
	})
	log.Println("Application has started. Listening port is 4500")
	err := http.ListenAndServe(":4500", nil)
	if err != nil {
		log.Fatal(err)
	}
}
