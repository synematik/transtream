package main

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"net/http"
)

const (
	width     = 640
	height    = 480
	frameSize = width * height * 4 // rgba: 4 bytes per pixel
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *State) StreamSocket(w http.ResponseWriter, r *http.Request) {
	log.WithField("socket", "connected").
		Error("Transcode")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	pr, pw := io.Pipe()

	transcoded := s.Transcode(pw, ManifestUrl, ffmpeg.KwArgs{
		//"vf": "scale=640:480",
	}, ffmpeg.KwArgs{
		"c:v":      "libx264",
		"preset":   "veryfast",
		"tune":     "zerolatency",
		"c:a":      "aac",
		"ar":       "44100",
		"movflags": "frag_keyframe+empty_moov+default_base_moof",
		"f":        "mp4",
		//"movflags": "frag_keyframe+empty_moov+default_base_moof+faststart",
		//"r":       "60",
	})

	buf := make([]byte, 4096) //32768

	for {
		n, err := pr.Read(buf)
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

		// Send the raw RGBA frame as a binary WebSocket message.
		err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		if err != nil {
			log.Println("Error writing to WebSocket:", err)
			break
		}
	}

	err = <-transcoded
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
		}).Error("Socket")
	}
	log.WithField("Socket", "").
		Info("closed ffmpeg pr")
}
