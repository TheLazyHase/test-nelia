package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"nelia/masker"
	"nelia/middleware"

	"github.com/esequiel378/hl7"
)

type GateHandler struct {
	Client *http.Client
	Filter *middleware.IngoingFilter
	Mask   *masker.Masker
	Gated  *string
}

func (g GateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	msg, err := hl7.ParseGeneric(data)
	if err == nil {
		//We need to apply HL7 rules
		err := g.Filter.ApplyHL7Rules(msg)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(err.Error()))
			return
		}
	}

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s%s", *g.Gated, r.RequestURI)

	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	proxyReq.Header = make(http.Header)
	for h, val := range r.Header {
		proxyReq.Header[h] = val
	}

	resp, err := g.Client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	gatedResult, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	returnedMessage, err := hl7.ParseGeneric(gatedResult)
	if err == nil {
		//We need to mask fields since it's a HL7 result
		returnedMessage, err := g.Mask.GetMaskedVersion(returnedMessage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		gatedResult = []byte(returnedMessage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(gatedResult)

	return
}
