package main

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// streamManifestHandler serves the HLS manifest file
func streamManifestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stream, err := streamManager.GetOrCreateStream(streamID, "http://example.com/source.mp4") // Replace with actual source URL
	if err != nil {
		http.Error(w, "Failed to create stream", http.StatusInternalServerError)
		return
	}

	manifestPath := filepath.Join(stream.TempDir, "manifest.m3u8")
	http.ServeFile(w, r, manifestPath)
}

// streamSegmentHandler serves individual HLS segments
func streamSegmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]
	segment := vars["segment"]

	stream, exists := streamManager.streams[streamID]
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	segmentPath := filepath.Join(stream.TempDir, segment)
	http.ServeFile(w, r, segmentPath)
}

// playHandler handles play requests
func playHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stream, exists := streamManager.streams[streamID]
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	stream.Play()
	w.WriteHeader(http.StatusOK)
}

// pauseHandler handles pause requests
func pauseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stream, exists := streamManager.streams[streamID]
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	stream.Pause()
	w.WriteHeader(http.StatusOK)
}

// seekHandler handles seek requests
func seekHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	timeParam := r.URL.Query().Get("time")
	position, err := strconv.ParseFloat(timeParam, 64)
	if err != nil {
		http.Error(w, "Invalid time parameter", http.StatusBadRequest)
		return
	}

	stream, exists := streamManager.streams[streamID]
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	if err := stream.Seek(position); err != nil {
		http.Error(w, "Failed to seek", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// wsHandler handles WebSocket connections for real-time updates
func wsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stream, exists := streamManager.streams[streamID]
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	stream.mu.Lock()
	stream.clients = append(stream.clients, conn)
	stream.mu.Unlock()

	// Send initial state
	conn.WriteJSON(map[string]interface{}{
		"type":      "state",
		"playing":   stream.IsPlaying,
		"position":  stream.getCurrentPosition(),
		"timestamp": time.Now().UnixMilli(),
	})

	// Handle WebSocket messages (e.g., client-side events)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// Remove client on disconnect
		stream.mu.Lock()
		for i, client := range stream.clients {
			if client == conn {
				stream.clients = append(stream.clients[:i], stream.clients[i+1:]...)
				break
			}
		}
		stream.mu.Unlock()
	}()
}
