package raft

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"sort"
	"sync"
	"time"
)

type LogEntry struct {
	Index   int
	Epoch   int
	Command string
	Args    MapCommandArgs
}

type Log struct {
	mu      sync.RWMutex
	Entries []LogEntry
}

func InitLog() *Log {
	initialLogEntry := LogEntry{Index: 0, Epoch: 0, Command: NoOp, Args: MapCommandArgs{}}
	return &Log{
		mu:      sync.RWMutex{},
		Entries: []LogEntry{initialLogEntry},
	}
}

func (l LogEntry) Matches(ol LogEntry) bool {
	return l.Index == ol.Index &&
		l.Command == ol.Command &&
		l.Args == ol.Args
}

func (l *Log) report() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return fmt.Sprintf("log: %v entries, last entry: %v... all entries: %v", len(l.Entries), l.Last(), l.Entries)
}

func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.Entries)
}

func (l *Log) Last() LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.Entries[len(l.Entries)-1]
}

func (l *Log) AddNewLogEntry(command string, args MapCommandArgs, epoch int) *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	last := l.Entries[len(l.Entries)-1]
	index := last.Index + 1
	entry := LogEntry{
		Index:   index,
		Epoch:   epoch,
		Command: command,
		Args:    args,
	}
	l.Entries = append(l.Entries, entry)

	return &entry
}

func (l *Log) Append(entries ...LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Entries = append(l.Entries, entries...)
}

func (l *Log) LogsBetween(low int, high int) []LogEntry {
	log.Printf("finding logs between %v and %v...", low, high)
	l.mu.RLock()
	defer l.mu.RUnlock()

	entries := make([]LogEntry, 0)
	count := len(l.Entries)
	low_index := sort.Search(count-1, func(i int) bool {
		log.Printf("entry index %v", l.Entries[i].Index)
		return l.Entries[i].Index > low
	})

	log.Printf("low index %v", low_index)

	for i := low_index; i < count; i++ {
		entry := l.Entries[i]
		if entry.Index > high || entry.Index < low {
			break
		}

		entries = append(entries, entry)
	}

	log.Printf("found %v logs between %v and %v...", len(entries), low, high)
	return entries
}

func (n *Node) minPushedIndex() int {
	n.cLock.Lock()
	defer n.cLock.Unlock()

	if len(n.PushedIndex) == 0 {
		return n.Log.Last().Index
	}

	min := math.MaxInt
	for _, pushed := range n.PushedIndex {
		if pushed < min {
			min = pushed
		}
	}
	return min
}

func (n *Node) commitLogs(maxToCommit int) {
	log.Printf("committing logs to %v...", maxToCommit)
	n.cLock.Lock()
	defer n.cLock.Unlock()

	nextToCommit := n.CommittedIndex + 1
	uncommitted := n.Log.LogsBetween(nextToCommit, maxToCommit)
	for _, entry := range uncommitted {
		reply, err := n.State.Apply(entry.Command, entry.Args)
		if err != nil {
			log.Printf("error committing logs: %v", err)
		}
		n.ClientService.notify(entry, reply)
	}

	n.CommittedIndex = maxToCommit
}

func (n *Node) pushLogsToFollower(addr net.TCPAddr) {
	log.Printf("pushLogsToFollower %v %v", addr.Network(), addr.String())
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("pushLogsToFollower error for %v: %v", addr, e)
		return
	}

	n.cLock.Lock()
	minPush := n.PushedIndex[addr.String()]
	n.cLock.Unlock()

	lastEntry := n.Log.Last()
	entries := n.Log.LogsBetween(minPush, lastEntry.Index)

	args := &AppendArgs{
		View:           *n.View,
		LastIndex:      minPush,
		Entries:        entries,
		CommittedIndex: n.CommittedIndex,
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
		return
	}

	if reply.Accepted {
		n.cLock.Lock()
		n.PushedIndex[addr.String()] = lastEntry.Index
		n.cLock.Unlock()
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
		log.Printf("epoch is old")
		reply.Accepted = false
		reply.View = *n.View

	case n.Log.Last().Index != args.LastIndex:
		log.Printf("last index doesn't match %v vs %v", n.Log.Last().Index, args.LastIndex)
		reply.Accepted = false
		reply.View = *n.View
		n.Pinged <- true

	case args.View.Epoch > n.View.Epoch:
		log.Printf("epoch is superior")
		reply.Accepted = true
		n.View = args.View.Clone()
		reply.View = *n.View
		n.Pinged <- true

	case args.View.Epoch == n.View.Epoch:
		log.Printf("epoch matches")
		reply.Accepted = true
		if !args.View.Same(*n.View) {
			log.Printf("received an append entries from same epoch but different view: \n mine: %v \n args: %v", n.View, args.View)
			n.View = args.View.Clone()
		}
		reply.View = *n.View
		n.Pinged <- true
	}

	if reply.Accepted {
		log.Printf("appending logs...")
		n.Log.Append(args.Entries...)

		log.Printf("commiting logs to %v...", args.CommittedIndex)
		n.commitLogs(args.CommittedIndex)
	}

	return nil
}
