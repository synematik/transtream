package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Stream represents a managed video stream
type Stream struct {
	ID         string
	SourceURL  string
	FFmpegCmd  *exec.Cmd
	TempDir    string
	IsPlaying  bool
	Position   float64
	lastUpdate time.Time
	clients    []*websocket.Conn
	mu         sync.Mutex
	startTime  time.Time
	pauseTime  time.Time
}

// StreamManager manages active streams
type StreamManager struct {
	streams map[string]*Stream
	mu      sync.Mutex
}

// NewStreamManager creates a new StreamManager
func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

// GetOrCreateStream retrieves or creates a new stream
func (sm *StreamManager) GetOrCreateStream(id, sourceURL string) (*Stream, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if stream, exists := sm.streams[id]; exists {
		return stream, nil
	}

	tempDir, err := os.MkdirTemp("", "stream_"+id)
	if err != nil {
		return nil, err
	}

	stream := &Stream{
		ID:        id,
		SourceURL: sourceURL,
		TempDir:   tempDir,
		IsPlaying: true,
		startTime: time.Now(),
	}

	if err := stream.startFFmpeg(); err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}

	sm.streams[id] = stream
	return stream, nil
}

// startFFmpeg initializes FFmpeg process for HLS streaming
func (s *Stream) startFFmpeg() error {
	args := []string{
		"-ss", fmt.Sprintf("%.2f", s.Position),
		"-i", s.SourceURL,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-hls_time", "2",
		"-hls_list_size", "3",
		"-hls_segment_filename", filepath.Join(s.TempDir, "segment_%03d.ts"),
		filepath.Join(s.TempDir, "manifest.m3u8"),
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	s.FFmpegCmd = cmd
	s.lastUpdate = time.Now()
	return nil
}

// Control methods with synchronization
func (s *Stream) Play() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsPlaying {
		return
	}

	if cmd := s.FFmpegCmd.Process; cmd != nil {
		cmd.Signal(syscall.SIGCONT)
	}

	s.startTime = time.Now().Add(-time.Duration(s.Position * float64(time.Second)))
	s.IsPlaying = true
	s.broadcastState()
}

func (s *Stream) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsPlaying {
		return
	}

	if cmd := s.FFmpegCmd.Process; cmd != nil {
		cmd.Signal(syscall.SIGSTOP)
	}

	s.Position = s.getCurrentPosition()
	s.IsPlaying = false
	s.broadcastState()
}

func (s *Stream) Seek(position float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Kill existing process
	if err := s.FFmpegCmd.Process.Kill(); err != nil {
		return err
	}

	// Reset state
	s.Position = position
	s.startTime = time.Now()
	s.IsPlaying = true

	// Restart FFmpeg with new position
	if err := s.startFFmpeg(); err != nil {
		return err
	}

	s.broadcastState()
	return nil
}

// Helper to calculate current position
func (s *Stream) getCurrentPosition() float64 {
	if s.IsPlaying {
		return time.Since(s.startTime).Seconds()
	}
	return s.Position
}

// Broadcast state to all WebSocket clients
func (s *Stream) broadcastState() {
	state := map[string]interface{}{
		"type":      "state",
		"playing":   s.IsPlaying,
		"position":  s.getCurrentPosition(),
		"timestamp": time.Now().UnixMilli(),
	}

	for _, client := range s.clients {
		if err := client.WriteJSON(state); err != nil {
			fmt.Println("WebSocket write error:", err)
		}
	}
}
