// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

// TranscodejobInterface is interface for /transcode/job root endpoint
type TranscodejobInterface interface { // Post is the handler for POST /transcode/job
	Post(http.ResponseWriter, *http.Request)
}

// TranscodejobInterfaceRoutes is routing for /transcode/job root endpoint
func TranscodejobInterfaceRoutes(r *mux.Router, i TranscodejobInterface) {
	r.HandleFunc("/transcode/job", i.Post).Methods("POST")
}
