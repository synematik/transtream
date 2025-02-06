package main

import (
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func read(input string, writer io.WriteCloser) <-chan error {
	log.Println("Starting ffmpeg process1")
	done := make(chan error)
	go func() {
		err := ffmpeg.
			Input(input).
			Output("pipe:", ffmpeg.KwArgs{
				//"format":  "rawvideo",
				//"pix_fmt": "rgb24",
				//"pix_fmt": "yuv420p",
			}).
			WithOutput(writer).
			Run()
		log.Println("ffmpeg process1 done")
		_ = writer.Close()
		done <- err
		close(done)
	}()
	return done
}

func Handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Disposition", "attachment; filename=video.mp4")
	writer.Header().Set("Content-Type", "application/octet-stream")

	pr, pw := io.Pipe()

	done := read("video.mp4", pw)

	_, err := io.Copy(writer, pr)
	if err != nil {
		http.Error(writer, "Failed to stream file", http.StatusInternalServerError)
	}

	err = <-done
	if err != nil {
		log.Println("FFmpeg error:", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	router.HandleFunc("/", Handler).
		Methods("GET")

	addr := "192.168.137.137:8080"
	log.Println("Serving: http://" + addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
