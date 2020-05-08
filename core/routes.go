package main

import (
	"net/http"
)

func getHello(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"value": "Hello"}`))

	return
}
