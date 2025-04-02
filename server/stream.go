package main

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"sync"
)

// Stream represents the shared state and distribution mechanism.
type Stream struct {
	// broadcast channel: the ffmpeg pipe will push data chunks here.
	broadcast chan []byte
	// register new client channels.
	register chan chan []byte
	// unregister client channels.
	unregister chan chan []byte
	// map to hold all subscriber channels.
	clients map[chan []byte]bool
	// lastData holds the latest chunk produced by ffmpeg.
	lastData []byte

	mu sync.RWMutex
}

func DefaultStream() *Stream {
	return &Stream{
		broadcast:  make(chan []byte, BroadcastHistoryCapacity),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		clients:    make(map[chan []byte]bool),
	}
}

// BroadcastRegistry listens for register/unregister and broadcast events.
func (s *Stream) BroadcastRegistry() {
	for {
		select {
		case client := <-s.register:
			log.WithField("register", client).Info("BroadcastRegistry")
			s.mu.Lock()
			s.clients[client] = true
			// When a client connects, immediately send the latest data if available.
			if s.lastData != nil {
				client <- s.lastData
			}
			s.mu.Unlock()
		case client := <-s.unregister:
			log.WithField("unregister", client).Info("BroadcastRegistry")
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client)
			}
			s.mu.Unlock()
		case message := <-s.broadcast:
			log.WithField("broadcast", len(message)).Trace("BroadcastRegistry")
			// Save the most recent message.
			s.mu.Lock()
			s.lastData = message
			// BroadcastRegistry to all clients.
			for client := range s.clients {
				// Non-blocking send: if a client’s buffer is full, you can choose to
				// disconnect the client or drop the message. Here we drop the message.
				select {
				case client <- message:
				default:
					// Optionally, handle slow consumers.
					log.Println("Dropping message for a slow client")
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Stream) Transcode(pw *io.PipeWriter) <-chan error {
	log.WithField("transcoding", "started").
		Info("Transcode")
	done := make(chan error)
	go func() {
		err := ffmpeg.
			Input(ManifestUrl, ffmpeg.KwArgs{
				"re": "",
			}).
			Output("pipe:1", ffmpeg.KwArgs{
				//"re":             "",
				"y":              "",
				"r":              "30",
				"g":              "1",
				"s":              "1920x1080",
				"quality":        "realtime",
				"speed":          "7",
				"threads":        "8",
				"row-mt":         "1",
				"qmin":           "4",
				"qmax":           "48",
				"tile-columns":   "2",
				"frame-parallel": "1",
				"b:v":            "6000k",
				"c:v":            "libvpx-vp9",
				"b:a":            "196k",
				"c:a":            "libopus",
				//"force_key_frames": "expr:gte(t,n_forced*5)",
				//"x264-params":      "keyint=20:scenecut=0",
				"f": "webm",
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

func (s *Stream) RegisterStream() {
	log.WithField("streaming", "started").
		Info("RegisterStream")

	pr, pw := io.Pipe()

	transcoded := s.Transcode(pw)

	reader := bufio.NewReader(pr)
	buf := make([]byte, TransBufSize)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.WithFields(log.Fields{
					"ok":    "EOF",
					"trace": err,
				}).Info("Transcode")
				break
			}
			log.WithFields(log.Fields{
				"on":  "reading from ffmpeg pipe",
				"err": err,
			}).Error("Transcode")
			break
		}
		// It’s important to copy the data since buf will be reused.
		chunk := make([]byte, n)
		copy(chunk, buf[:n])
		log.WithField("transcoded", n).
			Trace("RegisterStream")
		s.broadcast <- chunk
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
