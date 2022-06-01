package raft

import (
	"fmt"
	"log"
	"sync"
)

type ClientService struct {
	Node    *Node
	mu      sync.Mutex
	Waiting map[LogEntry]chan MapCommandResult
}

func InitClientService() *ClientService {
	return &ClientService{
		Node:    nil,
		mu:      sync.Mutex{},
		Waiting: make(map[LogEntry]chan MapCommandResult),
	}
}

func (cs *ClientService) report() string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return fmt.Sprintf("client service: %v waiting", len(cs.Waiting))
}

func (cs *ClientService) notify(logEntry LogEntry, result MapCommandResult) {
	log.Printf("notify %+v with %+v", logEntry, result)
	cs.mu.Lock()
	waiting, exists := cs.Waiting[logEntry]
	if !exists {
		log.Printf("notify %+v no one is waiting", logEntry)
		cs.mu.Unlock()
		return
	}
	cs.mu.Unlock()

	waiting <- result
	log.Printf("notified %+v with %+v", logEntry, result)
}

func (cs *ClientService) unregister(logEntry LogEntry) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.Waiting, logEntry)
}

func (cs *ClientService) register(logEntry LogEntry) chan MapCommandResult {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	waitChan := make(chan MapCommandResult)
	cs.Waiting[logEntry] = waitChan

	return waitChan
}

func (cs *ClientService) Request(args *ClientCommandArgs, reply *ClientCommandReply) error {
	log.Printf("received request %v", args)
	entry := cs.Node.Log.AddNewLogEntry(args.Command, args.Args, cs.Node.View.Epoch)

	waiting := cs.register(*entry)
	res := <-waiting
	cs.unregister(*entry)

	reply.Id = args.Id
	reply.Command = args.Command

	reply.Result = res
	reply.Error = nil
	reply.View = *cs.Node.View

	log.Printf("sending reply %v", reply)

	return nil
}
