package main

import (
	"io"
	"log"
	"net/http"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
	baseURL := "https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8"

	pr, pw := io.Pipe()

	go func() {
		err := ffmpeg.
			Input(baseURL).
			Output("pipe:1", ffmpeg.KwArgs{
				"c:v": "libx264", // Choose a codec like H.264
				"c:a": "aac",     // Audio codec
				"f":   "mp4",     // Output format (MP4)
			}).
			WithOutput(pw).
			Run()

		if err != nil {
			log.Println("FFmpeg error:", err)
		}
		pw.Close()
	}()

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	_, err := io.Copy(w, pr)
	if err != nil {
		log.Println("Failed to stream video:", err)
		http.Error(w, "Failed to stream video", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/stream", streamHandler)

	addr := ":8080"
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(addr, nil))
}
