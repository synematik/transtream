package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	state := NewState()

	go state.StartStream()

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)
	router.HandleFunc("/multi", state.StreamHandler).Methods("GET")
	router.HandleFunc("/state", state.StateHandler).Methods("GET")

	addr := "127.0.0.1:8080"
	log.Println("Serving: http://" + addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
