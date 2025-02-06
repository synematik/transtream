package main

import (
	"bufio"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
	"time"
)

// func (s *State) Transcode(pw *io.PipeWriter, manifestURL string) <-chan error {
func (s *State) Transcode(pw *io.PipeWriter, manifestURL string) {
	log.Println("Transcoding...")
	//done := make(chan error)
	go func() {
		//err := ffmpeg.
		ffmpeg.
			Input(manifestURL).
			Output("pipe:1", ffmpeg.KwArgs{
				"format":   "mp4",
				"vcodec":   "libx264",
				"preset":   "ultrafast",
				"tune":     "zerolatency",
				"movflags": "frag_keyframe+empty_moov+faststart",
			}).
			WithOutput(pw).
			Run()

		_ = pw.Close()
		//done <- err
		//close(done)
	}()
	//return done
}

func (s *State) Stream() {
	s.once.Do(func() {
		log.Println("Streaming...")

		pr, pw := io.Pipe()

		//transcoded := s.Transcode(pw, ManifestUrl)
		go s.Transcode(pw, ManifestUrl)
		//err := <-transcoded
		//if err != nil {
		//	panic(err)
		//}

		//broadcasted := s.Broadcast(pr)
		go s.Broadcast(pr)
		//err := <-broadcasted
		//if err != nil {
		//	panic(err)
		//}
	})
}

// func (s *State) Broadcast(pr *io.PipeReader) <-chan error {
func (s *State) Broadcast(pr *io.PipeReader) {
	log.Println("Broadcasting...")
	//done := make(chan error)
	go func() {
		reader := bufio.NewReader(pr)
		buf := make([]byte, 4096)

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

			s.clients.Range(func(key, _ interface{}) bool {
				pw := key.(*io.PipeWriter) // Get PipeWriter from sync.Map
				_, err := pw.Write(buf[:n])
				if err != nil {
					log.Println("Error writing to client:", err)
					s.RemoveClient(pw) // Remove client on error
				}
				return true // Continue iterating
			})
		}

		// Close pipe reader
		pr.Close()
	}()
	//return done
}
