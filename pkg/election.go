package raft

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

func (n *Node) requestNodeVote(addr net.TCPAddr, proposedView *RaftView, t time.Duration) bool {
	c, e := rpc.Dial(addr.Network(), addr.String())
	if e != nil {
		log.Printf("requestNodeVote error for %v: %v", addr, e)
		return false
	}

	args := &RequestVoteArgs{Proposed: *proposedView}
	reply := &RequestVoteReply{}
	call := c.Go("Node.RequestVote", args, reply, nil)
	select {
	case <-time.After(t):
		log.Printf("requestNodeVote timed out for %v", addr)
		return false

	case <-call.Done:
	}

	if call.Error != nil {
		log.Printf("requestNodeVote error for %v: %v", addr, call.Error)
		return false
	}

	return reply.Accepted
}

func (n *Node) requestNodeVoteWithTimeout(addr net.TCPAddr, proposedView *RaftView, result chan bool) {
	t, _ := time.ParseDuration("1s")
	result <- n.requestNodeVote(addr, proposedView, t)
}

func (n *Node) runElection() error {
	oldView := n.View

	votes := 1
	n.Role = Candidate
	n.View = oldView.IncrementView(n.Addr)
	results := make(chan bool)

	log.Printf("running election for epoch %v", n.View.Epoch)
	for _, a := range n.View.Followers {
		go n.requestNodeVoteWithTimeout(a, n.View, results)
	}

	for range n.View.Followers {
		res := <-results
		if res {
			votes += 1
		}
	}

	if votes < n.View.Quorum() {
		return fmt.Errorf("lost election %v", n.Addr.String())
	}

	log.Printf("won election %v", n.Addr.String())
	n.Role = Leader
	return nil
}

func (n *Node) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	switch {
	case args.Proposed.Epoch < n.View.Epoch:
		reply.Accepted = false
		log.Printf("received vote request for old epoch %v", n.View.Epoch)

	case args.Proposed.Epoch == n.View.Epoch:
		reply.Accepted = false
		log.Printf("received vote request for same epoch %v", n.View.Epoch)

	case args.Proposed.Epoch > n.View.Epoch:
		n.Role = Follower
		n.Pinged <- true
		reply.Accepted = true
		n.View = args.Proposed.Clone()
		log.Printf("received vote request for new epoch %v", n.View.Epoch)
	}

	return nil
}
