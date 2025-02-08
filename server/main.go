package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	ManifestUrl              string = "https://synema.cxllmerichie.com/proxy/f2c77277c3ae531faac9c32d2c04863d:2025020822:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydGdnRjhYVGZDZmlIYUYyWjU2eVRSZ0E9PQ==/2/4/8/5/0/7/2el8n.mp4:hls:manifest.m3u8"
	Address                  string = "0.0.0.0:8079"
	TransBufSize             uint   = 32 * 1024
	SockBufSize              uint   = 4 * 1024
	BroadcastHistoryCapacity        = 16
)

func state() *Stream {
	s := DefaultStream()

	go s.BroadcastRegistry()
	go s.RegisterStream()

	return s
}

func app() *mux.Router {
	r := mux.NewRouter()
	s := state()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.WithField("uri", r.RequestURI).Info("middleware")
			next.ServeHTTP(w, r)
		})
	})

	r.HandleFunc("/socket", s.SocketStreamHandler)
	r.HandleFunc("/", s.StreamHandler)
	r.HandleFunc("/stream", s.BroadcastHandler)

	return r
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		DisableQuote:           true,
		FullTimestamp:          true,
		TimestampFormat:        "15:04:05.000",
	})
	log.SetLevel(log.TraceLevel)

	log.WithField("server", "").
		Info("Serving: http://" + Address)
	log.WithField("server", "").
		Fatal(http.ListenAndServe(Address, app()))
}
