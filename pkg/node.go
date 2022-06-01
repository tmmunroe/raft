package raft

import (
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Role string

const (
	Leader    = Role("Leader")
	Candidate = Role("Candidate")
	Follower  = Role("Follower")
)

const (
	LeaderTick   = time.Duration(2 * time.Second)
	FollowerTick = time.Duration(7 * time.Second)

	CandidateSecMin = 3
	CandidateSecMax = 5
)

type Node struct {
	Role          Role
	Addr          net.TCPAddr
	View          *RaftView
	Log           *Log
	State         *MapState
	ClientService *ClientService

	cLock          sync.Mutex
	CommittedIndex int
	PushedIndex    map[string]int

	Pinged chan bool
}

func InitNode(addr net.TCPAddr, others []net.TCPAddr, state *MapState) *Node {
	pushed := make(map[string]int)
	for _, o := range others {
		pushed[o.String()] = 0
	}

	return &Node{
		Role:           Follower,
		Addr:           addr,
		Log:            InitLog(),
		View:           InitView(0, addr, others),
		State:          state,
		ClientService:  InitClientService(),
		cLock:          sync.Mutex{},
		CommittedIndex: -1,
		PushedIndex:    pushed,
		Pinged:         make(chan bool),
	}
}

func (n *Node) startServer() error {
	log.Printf("server starting %v", n.Addr.String())
	l, e := net.Listen(n.Addr.Network(), n.Addr.String())
	if e != nil {
		return e
	}

	a, e := net.ResolveTCPAddr(l.Addr().Network(), l.Addr().String())
	if e != nil {
		return e
	}
	n.Addr = *a

	s := rpc.NewServer()
	s.Register(n)
	s.Register(n.ClientService)
	n.ClientService.Node = n

	go s.Accept(l)
	return nil
}

func (n *Node) report() {
	log.Printf("%v %v report (lastCommitted: %v, pushed: %v)", n.Role, n.Addr.String(), n.CommittedIndex, n.PushedIndex)
	log.Print(n.View.report())
	log.Print(n.Log.report())
	log.Print(n.State.report())
	log.Print(n.ClientService.report())
}

func (n *Node) leaderTick() {
	log.Printf("leader ticking %v", n.Addr.String())
	n.report()
	select {
	case <-time.After(LeaderTick):
		n.pushLogsToFollowers()
		n.commitLogs(n.minPushedIndex())

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) candidateTick() {
	log.Printf("candidate ticking %v", n.Addr.String())

	randTick := RandomDuration(CandidateSecMin, CandidateSecMax)
	select {
	case <-time.After(randTick):
		log.Printf("node hasn't heard from leader")
		n.runElection()

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) followerTick() {
	log.Printf("follower ticking %v", n.Addr.String())

	select {
	case <-time.After(FollowerTick):
		log.Printf("node hasn't heard from leader")
		n.runElection()

	case <-n.Pinged:
		log.Printf("node heard from leader")
	}
}

func (n *Node) Run() error {
	log.Printf("node starting %v", n.Addr.String())

	e := n.startServer()
	if e != nil {
		return e
	}

	for {
		switch n.Role {
		case Leader:
			n.leaderTick()

		case Candidate:
			n.candidateTick()

		case Follower:
			n.followerTick()
		}

	}

}
