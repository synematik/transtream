package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	log.SetLevel(log.TraceLevel)

	state := NewState()

	go state.Stream(ManifestUrl)

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)
	router.HandleFunc("/", state.StreamHandler).Methods("GET")
	router.HandleFunc("/state", state.StateHandler).Methods("GET")

	log.WithField("server", "").Info("Serving: http://" + Address)
	log.WithField("server", "").Fatal(http.ListenAndServe(Address, router))
}
