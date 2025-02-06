package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	state := NewSharedState()

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)
	router.HandleFunc("/", state.StreamHandler)

	addr := "127.0.0.1:8080" //"192.168.137.137:8080"
	log.Println("Serving: http://" + addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
