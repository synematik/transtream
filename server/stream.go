package main

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"time"
)

//// func (s *State) Transcode(pw *io.PipeWriter, manifestURL string) <-chan error {
//func (s *State) Transcode(pw *io.PipeWriter, manifestURL string) {
//	log.Println("(?) transcoding...")
//	//done := make(chan error)
//	go func() {
//		err := ffmpeg.
//			Input(manifestURL).
//			Output("pipe:1", ffmpeg.KwArgs{
//				"format":   "mp4",
//				"vcodec":   "libx264",
//				"preset":   "ultrafast",
//				"tune":     "zerolatency",
//				"movflags": "frag_keyframe+empty_moov+faststart",
//			}).
//			WithOutput(pw).
//			Run()
//		if err != nil {
//			log.Println("(ERR) calling ffmpeg:", err)
//		}
//
//		err = pw.Close()
//		if err != nil {
//			log.Println("(ERR) closing ffmpeg pw:", err)
//		}
//		//done <- err
//		//close(done)
//	}()
//	//return done
//}
//
//func (s *State) Stream() {
//	s.once.Do(func() {
//		log.Println("[Stream] (?) streaming...")
//
//		pr, pw := io.Pipe()
//
//		//transcoded := s.Transcode(pw, ManifestUrl)
//		go s.Transcode(pw, ManifestUrl)
//		//err := <-transcoded
//		//if err != nil {
//		//	panic(err)
//		//}
//
//		//broadcasted := s.Broadcast(pr)
//		go s.Broadcast(pr)
//		//err := <-broadcasted
//		//if err != nil {
//		//	panic(err)
//		//}
//	})
//}
//
//// func (s *State) Broadcast(pr *io.PipeReader) <-chan error {
//func (s *State) Broadcast(pr *io.PipeReader) {
//	log.Println("(?) broadcasting...")
//	//done := make(chan error)
//	go func() {
//		reader := bufio.NewReader(pr)
//		buf := make([]byte, 4096)
//
//		for {
//			if !s.isActive {
//				timeout := 500 * time.Millisecond
//				log.Println("(?) blocked for", timeout, "ms...")
//				time.Sleep(timeout)
//				continue
//			}
//
//			n, err := reader.Read(buf)
//			log.Println("(OK) broadcast", n, "bytes")
//			if err != nil {
//				if err == io.EOF {
//					log.Println("(OK) completed.")
//					break
//				}
//				log.Println("(ERR) reading ffmpeg pr:", err)
//				break
//			}
//
//			s.clients.Range(func(client, _ interface{}) bool {
//				pw := client.(*io.PipeWriter)
//				n, err = pw.Write(buf[:n])
//				if err != nil {
//					log.Println("(ERR) sharing with", pw, "failed:", err)
//					s.RemoveClient(pw)
//				}
//				log.Println("(OK) shared", n, "bytes with", pw)
//				return true // Continue iterating
//			})
//		}
//
//		err := pr.Close()
//		if err != nil {
//			log.Println("(ERR) closing ffmpeg pr:", err)
//		}
//		log.Println("(OK) closed ffmpeg pr")
//	}()
//	//return done
//}

func (s *State) Transcode(pw *io.PipeWriter, manifestURL string) <-chan error {
	log.WithField("transcoding", "started").
		Info("Transcode")
	done := make(chan error)
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
			log.WithFields(log.Fields{
				"on":  "ffmpeg call",
				"err": err,
			}).Error("Transcode")
			done <- err
			close(done)
		}

		err = pw.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"on":  "closing ffmpeg pw",
				"err": err,
			}).Error("Transcode")
			done <- err
			close(done)
		}

		close(done)
	}()
	return done
}

func (s *State) Stream(manifestURL string) {
	s.once.Do(func() {
		log.WithField("streaming", "started").
			Info("Stream")

		pr, pw := io.Pipe()

		transcoded := s.Transcode(pw, manifestURL)

		reader := bufio.NewReader(pr)
		buf := make([]byte, 4096)

		for {
			var numClients uint8
			s.clients.Range(func(k, v interface{}) bool {
				numClients++
				return true
			})

			if !s.isActive || numClients == 0 {
				timeout := 500 * time.Millisecond
				log.WithField("timeout", timeout).
					Info("Broadcast")
				time.Sleep(timeout)
				continue
			}

			n, err := reader.Read(buf)
			log.WithField("reading", n).
				Trace("Broadcast")
			if err != nil {
				if err == io.EOF {
					log.WithField("completed", "EOF").
						Info("Broadcast")
					break
				}
				log.WithFields(log.Fields{
					"on":  "reading ffmpeg pr",
					"err": err,
				}).Error("Broadcast")
				break
			}

			s.clients.Range(func(client, _ interface{}) bool {
				pw := client.(*io.PipeWriter)
				n, err = pw.Write(buf[:n])
				if err != nil {
					log.WithFields(log.Fields{
						"on":  "sharing with pw",
						"err": err,
					}).Error("Broadcast")
					s.RemoveClient(pw)
				}
				log.WithField("shared", n).
					Trace("Broadcast")
				return true // Continue iterating
			})
		}

		err := <-transcoded
		if err != nil {
			log.WithField("err", err).
				Error("Transcode")
			panic(err)
		}
		err = pr.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"on":  "closing ffmpeg pr",
				"err": err,
			}).Error("Broadcast")
		}
		log.WithField("Broadcast", "").
			Info("closed ffmpeg pr")
	})
}
