package main

import (
	"bufio"
	"encoding/json"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
	"net/http"
	"time"
)

func (s *State) SingleStreamHandler(w http.ResponseWriter, r *http.Request) {
	//	w.Header().Set("Content-Type", "application/octet-stream")
	//	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "video/mp4")
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

	// Start FFmpeg process
	go func() {
		err := ffmpeg.
			Input(ManifestUrl).
			Output("pipe:1", ffmpeg.KwArgs{
				//"listen":   "1",
				"f":        "mp4",
				"vcodec":   "libx264",
				"preset":   "ultrafast",
				"tune":     "zerolatency",
				"movflags": "frag_keyframe+empty_moov+faststart",
				"filter:v": "fps=60",
			}).
			WithOutput(pw). // Write to pipe
			Run()

		if err != nil {
			log.Println("FFmpeg error:", err)
		}
		pw.Close() // Close pipe when FFmpeg finishes
	}()

	// Buffer for reading chunks
	reader := bufio.NewReader(pr)
	buf := make([]byte, 4096) // Adjust buffer size for better performance

	// Read from pipe and write to response
	for {
		if !s.isActive {
			log.Println("Blocked for 500ms...")
			time.Sleep(500 * time.Millisecond)
			continue
		}

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

	// Close pipe reader
	pr.Close()
}

type StateRequest struct {
	State bool    `json:"state"`
	Time  float64 `json:"time"`
}

func (s *State) SingleStateHandler(w http.ResponseWriter, r *http.Request) {
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
