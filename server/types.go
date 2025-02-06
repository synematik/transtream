package main

import (
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
