package raft

import (
	"fmt"
	"log"
	"sync"
)

type State interface {
	Apply(command string, args interface{}) (interface{}, error)
}

const (
	SetValue string = "SetValue"
	GetValue string = "GetValue"
)

type MapCommandArgs struct {
	Key   string
	Value string
}

type MapCommandResult struct {
	Key   string
	Value string
}

type MapState struct {
	mu    sync.RWMutex
	State map[string]string
}

func InitMapState() *MapState {
	return &MapState{
		mu:    sync.RWMutex{},
		State: make(map[string]string),
	}
}

func (ms *MapState) get(key string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	res, exists := ms.State[key]
	if !exists {
		return "", fmt.Errorf("key does not exist")
	}

	return res, nil
}

func (ms *MapState) set(key string, value string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.State[key] = value
	return nil
}

func (ms *MapState) Apply(command string, args interface{}) (interface{}, error) {
	log.Printf("applying %v args %v", command, args)
	mcArgs := args.(MapCommandArgs)
	k, v := mcArgs.Key, mcArgs.Value

	switch command {
	case SetValue:
		e := ms.set(k, v)
		if e != nil {
			return nil, e
		}
		return MapCommandResult{Key: k, Value: v}, nil

	case GetValue:
		r, e := ms.get(k)
		if e != nil {
			return nil, e
		}
		return MapCommandResult{Key: k, Value: r}, nil
	}

	return nil, nil
}
