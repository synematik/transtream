package main

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (s *Stream) StreamHandler(w http.ResponseWriter, r *http.Request) {
	//	w.Header().Set("Content-Type", "application/octet-stream")
	//	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	pr, pw := io.Pipe()

	transcoded := s.Transcode(pw)

	// Buffer for reading chunks
	reader := bufio.NewReader(pr)
	buf := make([]byte, TransBufSize) // Adjust buffer size for better performance

	// Read from pipe and write to response
	for {
		n, err := reader.Read(buf)
		log.Println(n, "bytes chunk")
		if err != nil {
			if err == io.EOF {
				break // FFmpeg process ended
			}
			log.Println("Error reading from FFmpeg:", err)
			break
		}

		// Write chunk to client
		_, err = w.Write(buf[:n])
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}

		// Flush to send data immediately
		flusher.Flush()
	}

	err := <-transcoded
	if err != nil {
		log.WithField("err", err).
			Error("Transcode")
	}

	err = pr.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"on":  "closing ffmpeg pr",
			"err": err,
		}).Error("BroadcastRegistry")
	}

	log.WithField("ok", "closed ffmpeg pr").
		Info("BroadcastRegistry")
}
