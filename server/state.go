package main

import (
	"sync"
)

type State struct {
	//mu       sync.Mutex
	clients  map[chan []byte]struct{}
	isActive bool
	once     sync.Once
}

func NewState() *State {
	return &State{
		clients:  make(map[chan []byte]struct{}),
		isActive: true,
	}
}
