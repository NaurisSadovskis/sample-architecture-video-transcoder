// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package main

import (
	"log"
	"net/http"

	"github.com/sample-architecture-video-transcoder/api/goraml"

	"github.com/gorilla/mux"
	"gopkg.in/validator.v2"
)

func main() {
	// input validator
	validator.SetValidationFunc("multipleOf", goraml.MultipleOf)

	r := mux.NewRouter()

	// home page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// apidocs
	r.PathPrefix("/apidocs/").Handler(http.StripPrefix("/apidocs/", http.FileServer(http.Dir("./apidocs/"))))

	initRoutes(r)

	log.Println("starting server")
	http.ListenAndServe(":5000", r)

}
