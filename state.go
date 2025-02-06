package main

import (
	"io"
	"sync"
)

type State struct {
	mu      sync.Mutex                // Ensures safe concurrent access
	streams map[string]*io.PipeReader // Active FFmpeg streams
}

func NewSharedState() *State {
	return &State{
		streams: make(map[string]*io.PipeReader),
	}
}
