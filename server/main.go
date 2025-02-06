package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		DisableQuote:           true,
		TimestampFormat:        "15:04:05.000000000",
		FullTimestamp:          true,
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
