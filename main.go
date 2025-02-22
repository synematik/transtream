package main

import (
	"bufio"
	//"fmt"
	"io"
	"log"
	"net/http"
	//"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const input string = "https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8"

func stream(w http.ResponseWriter, r *http.Request) {
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

	// Create a pipe for streaming FFmpeg output
	pr, pw := io.Pipe()

	// Start FFmpeg process
	go func() {
		err := ffmpeg.
			Input(input).
			Output("pipe:1", ffmpeg.KwArgs{
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
		if err != nil {
			if err == io.EOF {
				break // FFmpeg process ended
			}
			log.Println("Error reading from FFmpeg:", err)
			break
		}

		// Write chunk to client
		_, writeErr := w.Write(buf[:n])
		if writeErr != nil {
			log.Println("Client disconnected:", writeErr)
			break
		}

		// Flush to send data immediately
		flusher.Flush()
	}

	// Close pipe reader
	pr.Close()
}

//func stream(w http.ResponseWriter, r *http.Request) {
//	flusher, ok := w.(http.Flusher)
//	if !ok {
//		panic("expected http.ResponseWriter to be an http.Flusher")
//	}
//
//	w.Header().Set("Transfer-Encoding", "chunked")
//	w.Header().Set("Content-Type", "application/octet-stream")
//	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
//	w.Header().Set("Connection", "keep-alive")
//	w.Header().Set("Access-Control-Allow-Origin", "*")
//	w.Header().Set("X-Content-Type-Options", "nosniff")
//	for i := 1; i <= 10; i++ {
//		fmt.Fprintf(w, "Chunk #%d\n", i)
//		flusher.Flush() // Trigger "chunked" encoding and send a chunk...
//		time.Sleep(500 * time.Millisecond)
//	}
//
//
//	pr, pw := io.Pipe()
//
//	go func() {
//		err := ffmpeg.
//			Input(input).
//			Output("pipe", ffmpeg.KwArgs{
//				"listen":   "1",
//				"filter:v": "fps=60",
//				"f":        "mp4",
//				"preset":   "ultrafast",
//				"vcodec":   "libx264",
//				"tune":     "zerolatency",
//				"movflags": "frag_keyframe+empty_moov+faststart",
//			}).
//			WithOutput(pw).
//			Run()
//
//		if err != nil {
//			log.Println("FFmpeg error:", err)
//		}
//		pw.Close()
//	}()
//
//	_, err := io.Copy(w, pr)
//	if err != nil {
//		log.Println("Failed to stream video:", err)
//		http.Error(w, "Failed to stream video", http.StatusInternalServerError)
//	}
//}

func main() {
	http.HandleFunc("/", stream)

	addr := "127.0.0.1:8000"
	log.Println("Serving at: http://" + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
