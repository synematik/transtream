package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	state := NewState()

	go state.Stream()

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)
	router.HandleFunc("/", state.StreamHandler).Methods("GET")
	router.HandleFunc("/state", state.StateHandler).Methods("GET")

	log.Println("Serving: http://" + Address)
	log.Fatal(http.ListenAndServe(Address, router))
}
