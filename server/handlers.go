package main

import (
	"bufio"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (s *State) StreamHandler(w http.ResponseWriter, r *http.Request) {
	log.WithField("client", "connected").
		Info("StreamHandler")

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.WithFields(log.Fields{
			"on":  "flusher, ok := w.(http.Flusher)",
			"err": "unsupported keep-alive streaming ",
		}).Error("StreamHandler")
		http.Error(w, "HTTP keep-alive streaming unsupported", http.StatusInternalServerError)
		return
	}
	pr := s.AddClient() // Each client gets its own pipe reader
	//defer s.RemoveClient(pr)

	reader := bufio.NewReader(pr)
	buf := make([]byte, 2048)

	for {
		select {
		case <-r.Context().Done(): // Detect client disconnect
			log.WithField("client", "disconnected").
				Info("StreamHandler")
			return
		default:
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.WithField("StreamHandler", "").
						Trace("file ended, trace:", err)
					break
				}
				log.
					WithField("StreamHandler", "").
					Error("reading buffer:", err)
				break
			}
			log.WithField("read", n).
				Trace("StreamHandler")

			n, err = w.Write(buf[:n])
			if err != nil {
				log.WithFields(log.Fields{
					"on":  "writing buffer",
					"err": err,
				}).Error("StreamHandler")
				break
			}
			log.WithField("delivering", n).
				Trace("StreamHandler")

			flusher.Flush()
			log.WithField("delivered", n).
				Trace("StreamHandler")
		}
	}

	pr.Close()
}

func (s *State) StateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"on":  "request",
			"err": "failed to read request body",
		}).Error("StateHandler")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var req StateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.WithFields(log.Fields{
			"on":  "parsing JSON",
			"err": "invalid body",
		}).Error("StateHandler")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	s.isActive = req.State

	w.WriteHeader(http.StatusOK)
}
