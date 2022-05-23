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

func InitLog() *Log {
	return &Log{
		Entries: make([]LogEntry, 0),
	}
}

func (l LogEntry) Matches(ol LogEntry) bool {
	return l.Index == ol.Index &&
		l.Command == ol.Command &&
		l.Args == ol.Args
}

func (n *Node) pushLogsToFollower(addr net.TCPAddr) {
	log.Printf("pushLogsToFollower %v %v", addr.Network(), addr.String())
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("pushLogsToFollower error for %v: %v", addr, e)
		return
	}

	args := &AppendArgs{
		View:    *n.View,
		Entries: n.Log.Entries,
	}
	reply := &AppendReply{}

	log.Printf("pushLogsToFollower client %v args %v, reply %v", c, args, reply)
	call := c.Go("Node.AppendEntries", args, reply, nil)

	t := time.Duration(1 * time.Second)
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
	log.Printf("pushing logs to followers... %v", n.View)
	for _, addr := range n.View.Followers {
		go n.pushLogsToFollower(addr)
	}

	return nil
}

func (n *Node) AppendEntries(args *AppendArgs, reply *AppendReply) error {
	log.Printf("received append entries from view\n %v \n %v", args.View, n.View)
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
