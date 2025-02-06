package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithField("middleware", "").Info(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
