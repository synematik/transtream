package main

import (
	"io"
	"sync"
)

type StateRequest struct {
	State bool    `json:"state"`
	Time  float64 `json:"time"`
}

type State struct {
	clients  sync.Map
	isActive bool
	once     sync.Once
}

func NewState() *State {
	return &State{
		isActive: true,
	}
}

func (s *State) AddClient() *io.PipeReader {
	pr, pw := io.Pipe()
	s.clients.Store(pw, struct{}{})
	return pr
}

func (s *State) RemoveClient(pw *io.PipeWriter) {
	s.clients.Delete(pw)
	_ = pw.Close()
}
