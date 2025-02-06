package main

import (
	"sync"
)

type State struct {
	mu sync.Mutex
	//streams map[string]*io.PipeReader
	active bool
}

func NewState() *State {
	return &State{
		//streams: make(map[string]*io.PipeReader),
		active: true,
	}
}
