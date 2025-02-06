package main

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
)

func (s *State) StartStream() {
	s.once.Do(func() { // Ensures FFmpeg runs only once
		log.Println("Starting FFmpeg process...")

		// Create a pipe for streaming FFmpeg output
		pr, pw := io.Pipe()

		// Run FFmpeg in a separate goroutine
		go func() {
			err := ffmpeg.
				Input(ManifestUrl). // Replace with your HLS URL
				Output("pipe:1", ffmpeg.KwArgs{
					"format":   "mp4",
					"vcodec":   "libx264",
					"preset":   "ultrafast",
					"tune":     "zerolatency",
					"movflags": "frag_keyframe+empty_moov+faststart",
				}).
				WithOutput(pw).
				Run()

			if err != nil {
				log.Println("FFmpeg error:", err)
			}
			pw.Close() // Close writer when FFmpeg exits
		}()

		// Broadcast FFmpeg output to all connected clients
		go func() {
			buffer := make([]byte, 4096) // Chunk size
			for {
				n, err := pr.Read(buffer)
				if err != nil {
					log.Println("Stream read error:", err)
					break
				}

				// Send chunk to all connected clients
				s.mu.Lock()
				for ch := range s.clients {
					log.Println(ch)
					select {
					case ch <- buffer[:n]: // Send video data
					default: // Skip if channel is full
					}
				}
				s.mu.Unlock()
			}

			// Cleanup on stream end
			pr.Close()
			log.Println("FFmpeg stream ended.")
		}()
	})
}
