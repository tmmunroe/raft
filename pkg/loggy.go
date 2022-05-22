package raft

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

type LogEntry struct {
	Index   int
	Command string
	Args    interface{}
}

type Log struct {
	Entries []LogEntry
}

func (l LogEntry) Matches(ol LogEntry) bool {
	return l.Index == ol.Index &&
		l.Command == ol.Command &&
		l.Args == ol.Args
}

func (n *Node) pushLogsToFollower(addr *net.TCPAddr) {
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("pushLogsToFollower error for %v: %v", addr, e)
	}

	args := &AppendArgs{}
	reply := &AppendReply{}
	call := c.Go("Node.AppendEntries", args, reply, nil)

	t, _ := time.ParseDuration("1s")
	select {
	case <-time.After(t):
		log.Printf("pushLogsToFollower timed out for %v", addr)

	case <-call.Done:
	}

	if call.Error != nil {
		log.Printf("pushLogsToFollower error for %v: %v", addr, call.Error)
	}
}

func (n *Node) pushLogsToFollowers() error {
	for _, addr := range n.View.Followers {
		go n.pushLogsToFollower(addr)
	}

	return nil
}

func (n *Node) AppendEntries(args *AppendArgs, reply *AppendReply) error {
	switch {
	case args.View.Epoch < n.View.Epoch:
		reply.Accepted = false
		reply.View = *n.View

	case args.View.Epoch > n.View.Epoch:
		reply.Accepted = true
		n.View = args.View.Clone()
		reply.View = *n.View
		n.Pinged <- true

	case args.View.Epoch == n.View.Epoch:
		reply.Accepted = true
		if !args.View.Same(*n.View) {
			log.Printf("received an append entries from same epoch but different view: \n mine: %v \n args: %v", n.View, args.View)
			n.View = args.View.Clone()
		}
		n.Pinged <- true
	}

	return nil
}
