package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	raft "github.com/tmmunroe/raft/pkg"
)

func main() {
	var server int
	flag.IntVar(&server, "s", 0, "server number")
	flag.Parse()

	log.Printf("selected server %v", server)
	rand.Seed(time.Now().UnixNano())

	if server == -1 {
		log.Printf("server should be between 0 and 4")
		return
	}

	s := raft.InitMapState()
	addr, others := raft.SplitAddresses(server)
	others = make([]net.TCPAddr, 0)
	n := raft.InitNode(addr, others, s)
	e := n.Run()

	log.Printf("node stopped with error: %v", e)
}
