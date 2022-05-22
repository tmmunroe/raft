package raft

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

type Role string

const (
	Leader    = Role("Leader")
	Candidate = Role("Candidate")
	Follower  = Role("Follower")
)

type Node struct {
	Role   Role
	Addr   *net.TCPAddr
	View   *RaftView
	Log    *Log
	Pinged chan bool
}

func InitNode(addr *net.TCPAddr) *Node {
	return &Node{
		Role:   Follower,
		Addr:   addr,
		Log:    nil,
		View:   nil,
		Pinged: make(chan bool),
	}
}

func (n *Node) startServer() error {
	l, e := net.Listen(n.Addr.Network(), n.Addr.String())
	if e != nil {
		return e
	}

	n.Addr, e = net.ResolveTCPAddr(l.Addr().Network(), l.Addr().Network())
	if e != nil {
		return e
	}

	s := rpc.NewServer()
	s.Register(n)
	go s.Accept(l)
	return nil
}

func (n *Node) leaderTick() {
	tick, _ := time.ParseDuration("3s")

	select {
	case <-time.After(tick):
		n.pushLogsToFollowers()

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) candidateTick() {
	tick, _ := time.ParseDuration("5s")

	select {
	case <-time.After(tick):
		log.Printf("node hasn't heard from leader")
		n.runElection()

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) followerTick() {
	tick, _ := time.ParseDuration("5s")

	select {
	case <-time.After(tick):
		log.Printf("node hasn't heard from leader")
		n.runElection()

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) run() error {
	e := n.startServer()
	if e != nil {
		return e
	}

	for {
		switch n.Role {
		case Leader:
			n.leaderTick()

		case Follower:
			n.followerTick()

		case Candidate:
			n.candidateTick()
		}

	}

}
