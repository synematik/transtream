package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (s *State) StreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientChan := make(chan []byte, 4096)
	s.clients.Store(clientChan, struct{}{})

	// Ensure client is removed on disconnect
	defer func() {
		s.clients.Delete(clientChan)
		close(clientChan)
		log.Println("Client disconnected.")
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	reader.Re // Send chunks to client
	for chunk := range clientChan {
		_, err := w.Write(chunk)
		if err != nil {
			log.Println("Client write error:", err)
			break
		}
		flusher.Flush() // Ensure immediate send
	}
}

func (s *State) StateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var req StateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	s.isActive = req.State

	w.WriteHeader(http.StatusOK)
}
