package main

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func (s *State) StreamSocket(w http.ResponseWriter, r *http.Request) {
	log.WithField("socket", "connected").
		Error("Socket")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"on":  "upgrade",
			"err": err,
		}).Error("Socket")
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
		"movflags": "frag_keyframe+empty_moov+default_base_moof+faststart",
		"f":        "mp4",
		//"r":       "60",
	})

	buf := make([]byte, 4096) //32768

	for {
		n, err := pr.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.WithField("completed", "EOF").
					Info("Socket")
				break
			}
			log.WithFields(log.Fields{
				"on":  "reading ffmpeg pr",
				"err": err,
			}).Error("Socket")
			break
		}

		err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		if err != nil {
			log.WithFields(log.Fields{
				"on":  "conn.WriteMessage(websocket.BinaryMessage, buf[:n])",
				"err": err,
			}).Error("Socket")
			break
		}
	}
	err = <-transcoded
	if err != nil {
		log.WithField("err", err).
			Error("Socket")
	}
	err = pr.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"on":  "closing ffmpeg pr",
			"err": err,
		}).Error("Socket")
	}
	log.WithField("ok", "closed ffmpeg pr").
		Info("Socket")
}
