package main

import (
	"bufio"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (s *State) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.
		WithField("StreamHandler", "").
		Info("connected")

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.
			WithField("StreamHandler", "").
			Info("HTTP keep-alive streaming unsupported")
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
			log.WithField("StreamHandler", "").Info("Client disconnected.")
			return
		default:
			log.WithField("StreamHandler", "").Debug("Reading buffer...")
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.WithField("StreamHandler", "").Debug("file ended, trace:", err)
					break
				}
				log.
					WithField("StreamHandler", "").
					Error("reading buffer:", err)
				break
			}
			log.WithField("StreamHandler", "").Debug("read", n, "buffer bytes")

			n, err = w.Write(buf[:n])
			if err != nil {
				log.WithField("StreamHandler", "").Error("(ERR) writing buffer:", err)
				break
			}
			log.WithField("StreamHandler", "").Debug("delivering", n, "bytes...")

			flusher.Flush()
			log.WithField("StreamHandler", "").Debug("delivered", n, "bytes")
		}
	}

	pr.Close()
}

func (s *State) StateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithField("StateHandler", "").Error("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var req StateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.WithField("StateHandler", "").Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	s.isActive = req.State

	w.WriteHeader(http.StatusOK)
}
