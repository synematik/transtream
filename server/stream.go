package main

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
)

func (s *State) StartStream() {
	s.once.Do(func() { // Ensures FFmpeg runs only once
		log.Println("Starting FFmpeg process...")

		// Create pipe for FFmpeg output
		pr, pw := io.Pipe()

		// Start FFmpeg process
		go func() {
			err := ffmpeg.
				Input(ManifestUrl). // Your HLS source
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
			pw.Close()
		}()

		// Broadcast stream to clients
		go func() {
			buffer := make([]byte, 4096) // Chunk size
			for {
				n, err := pr.Read(buffer)
				if err != nil {
					log.Println("Stream read error:", err)
					break
				}

				// Send chunk to all connected clients
				s.clients.Range(func(key, value interface{}) bool {
					client := key.(chan []byte)
					log.Println(client)
					select {
					case client <- buffer[:n]: // Send video data
					default: // Skip if the client buffer is full
					}
					return true
				})
			}

			// Cleanup when stream ends
			pr.Close()
			log.Println("FFmpeg stream ended.")
		}()
	})
}
