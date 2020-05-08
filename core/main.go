package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("----started----")

	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/hello", getHello).Methods(http.MethodGet)

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO return all possible APIs
		fmt.Fprintln(w, "Welcome to v1 api")
	})

	log.Fatalln(http.ListenAndServe(":8080", r))
}
