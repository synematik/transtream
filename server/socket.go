package main

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Stream) SocketStreamHandler(w http.ResponseWriter, r *http.Request) {
	log.WithField("socket", "connected").
		Info("Socket")

	sock, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"on":  "upgrade",
			"err": err,
		}).Error("Socket")
		return
	}
	defer func(sock *websocket.Conn) {
		err := sock.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"on":  "sock.Close",
				"err": err,
			}).Error("Socket")
		}
	}(sock)

	// Create a channel for this client.
	stream := make(chan []byte, SockBufSize)
	// Register this client with the hub.
	s.register <- stream
	defer func() {
		s.unregister <- stream
	}()

	for chunk := range stream {
		err = sock.WriteMessage(websocket.BinaryMessage, chunk)
		log.WithField("sent", len(chunk)).
			Trace("Socket")
		if err != nil {
			log.WithFields(log.Fields{
				"on":  "sock.WriteMessage",
				"err": err,
			}).Error("Socket")
			break
		}
	}
}
