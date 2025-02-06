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
	log.Println("[Transcode] (?) transcoding...")
	//done := make(chan error)
	go func() {
		err := ffmpeg.
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
		if err != nil {
			log.Println("[Transcode] (ERR) calling ffmpeg:", err)
		}

		err = pw.Close()
		if err != nil {
			log.Println("[Transcode] (ERR) closing ffmpeg pw:", err)
		}
		//done <- err
		//close(done)
	}()
	//return done
}

func (s *State) Stream() {
	s.once.Do(func() {
		log.Println("[Stream] (?) streaming...")

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
	log.Println("[Broadcast] (?) broadcasting...")
	//done := make(chan error)
	go func() {
		reader := bufio.NewReader(pr)
		buf := make([]byte, 4096)

		for {
			if !s.isActive {
				timeout := 500 * time.Millisecond
				log.Println("[Broadcast] (?) blocked for", timeout, "ms...")
				time.Sleep(timeout)
				continue
			}

			n, err := reader.Read(buf)
			log.Println("[Broadcast] (OK) broadcast", n, "bytes")
			if err != nil {
				if err == io.EOF {
					log.Println("[Broadcast] (OK) completed.")
					break
				}
				log.Println("[Broadcast] (ERR) reading ffmpeg pr:", err)
				break
			}

			s.clients.Range(func(client, _ interface{}) bool {
				pw := client.(*io.PipeWriter)
				n, err = pw.Write(buf[:n])
				if err != nil {
					log.Println("[Broadcast] (ERR) sharing with", pw, "failed:", err)
					s.RemoveClient(pw)
				}
				log.Println("[Broadcast] (OK) shared", n, "bytes with", pw)
				return true // Continue iterating
			})
		}

		err := pr.Close()
		if err != nil {
			log.Println("[Broadcast] (ERR) closing ffmpeg pr:", err)
		}
		log.Println("[Broadcast] (OK) closed ffmpeg pr")
	}()
	//return done
}
