package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (s *State) MultiStreamHandler(w http.ResponseWriter, r *http.Request) {
	//if !s.isActive {
	//	http.Error(w, "Streaming is disabled", http.StatusForbidden)
	//	return
	//}

	//	w.Header().Set("Content-Type", "application/octet-stream")
	//	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan []byte, 10)
	s.mu.Lock()
	s.clients[clientChan] = struct{}{}
	s.mu.Unlock()

	// Ensure client is removed when disconnected
	defer func() {
		s.mu.Lock()
		delete(s.clients, clientChan)
		s.mu.Unlock()
		close(clientChan)
		log.Println("Client disconnected.")
	}()

	// Send chunks to the client
	for chunk := range clientChan {
		_, err := w.Write(chunk)
		if err != nil {
			log.Println("Client write error:", err)
			break
		}
		w.(http.Flusher).Flush() // Ensure the chunk is sent immediately
	}
}

func (s *State) MultiStateHandler(w http.ResponseWriter, r *http.Request) {
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

	s.mu.Lock()
	s.isActive = req.State
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}
