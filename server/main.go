package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	ManifestUrl              string = "https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8"
	Address                  string = "0.0.0.0:8079"
	TransBufSize             uint   = 32768
	SockBufSize              uint   = 4096
	BroadcastHistoryCapacity        = 16
)

func state() *Stream {
	s := DefaultStream()

	go s.BroadcastRegistry()
	go s.StreamSource(ManifestUrl)

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

	r.HandleFunc("/", s.SocketStreamHandler)

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
