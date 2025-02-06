package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Global stream manager to handle concurrent streams
var streamManager = NewStreamManager()

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	r := mux.NewRouter()

	// HTTP endpoints
	r.HandleFunc("/stream/{id}/manifest.m3u8", streamManifestHandler).Methods("GET")
	r.HandleFunc("/stream/{id}/{segment}", streamSegmentHandler).Methods("GET")
	r.HandleFunc("/api/play/{id}", playHandler).Methods("POST")
	r.HandleFunc("/api/pause/{id}", pauseHandler).Methods("POST")
	r.HandleFunc("/api/seek/{id}", seekHandler).Methods("POST")
	r.HandleFunc("/ws/{id}", wsHandler)

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
