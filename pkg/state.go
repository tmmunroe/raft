package raft

import (
	"fmt"
	"log"
	"sync"
)

const (
	NoOp     string = "NoOp"
	PutValue string = "PutValue"
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

func (ms *MapState) report() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return fmt.Sprintf("state: %v", ms.State)
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

func (ms *MapState) Clone() *MapState {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ns := InitMapState()
	for k, v := range ms.State {
		ns.State[k] = v
	}
	return ns
}

func (ms *MapState) Apply(command string, args MapCommandArgs) (MapCommandResult, error) {
	log.Printf("applying %v args %v", command, args)
	k, v := args.Key, args.Value

	switch command {
	case PutValue:
		e := ms.set(k, v)
		if e != nil {
			return MapCommandResult{}, e
		}
		return MapCommandResult{Key: k, Value: v}, nil

	case GetValue:
		r, e := ms.get(k)
		if e != nil {
			return MapCommandResult{}, e
		}
		return MapCommandResult{Key: k, Value: r}, nil

	case NoOp:
		return MapCommandResult{}, nil

	default:
		return MapCommandResult{}, fmt.Errorf("unknown command %v", command)
	}
}
