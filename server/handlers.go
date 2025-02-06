package main

import (
	"bufio"
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

	pr := s.AddClient() // Each client gets its own pipe reader
	//defer s.RemoveClient(pr)

	reader := bufio.NewReader(pr)
	buf := make([]byte, 4096)

	notify := r.Context().Done() // Detect client disconnect

	for {
		select {
		case <-notify:
			log.Println("Client disconnected.")
			return
		default:
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("Read error:", err)
				break
			}

			_, err = w.Write(buf[:n])
			if err != nil {
				log.Println("Write error:", err)
				break
			}

			w.(http.Flusher).Flush()
		}
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
