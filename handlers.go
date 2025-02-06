package main

import (
	"bufio"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
	"net/http"
)

func (s *State) StreamHandler(w http.ResponseWriter, r *http.Request) {
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
		n, err := reader.Read(buf)
		log.Println(n)
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
