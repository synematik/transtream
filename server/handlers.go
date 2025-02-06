package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (s *State) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[StreamHandler] connected")

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	//w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "HTTP keep-alive streaming unsupported", http.StatusInternalServerError)
		return
	}
	pr := s.AddClient() // Each client gets its own pipe reader
	//defer s.RemoveClient(pr)

	reader := bufio.NewReader(pr)
	buf := make([]byte, 4096)

	for {
		select {
		case <-r.Context().Done(): // Detect client disconnect
			log.Println("[StreamHandler] Client disconnected.")
			return
		default:
			log.Println("[StreamHandler] (?) Reading buffer...")
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("[StreamHandler] (OK) file ended, trace:", err)
					break
				}
				log.Println("[StreamHandler] (ERR) reading buffer:", err)
				break
			}
			log.Println("[StreamHandler] (OK) read", n, "buffer bytes")

			n, err = w.Write(buf[:n])
			if err != nil {
				log.Println("[StreamHandler] (ERR) writing buffer:", err)
				break
			}
			log.Println("[StreamHandler] (?) delivering", n, "bytes...")

			flusher.Flush()
			log.Println("[StreamHandler] (OK) delivered", n, "bytes")
		}
	}

	pr.Close()
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
