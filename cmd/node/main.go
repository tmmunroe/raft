package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	raft "github.com/tmmunroe/raft/pkg"
)

var addresses = []string{
	"127.0.0.1:51001",
	"127.0.0.1:51003",
	"127.0.0.1:51005",
	"127.0.0.1:51007",
	"127.0.0.1:51009",
}

func splitAddresses(server int) (net.TCPAddr, []net.TCPAddr) {
	a := addresses[server]

	addr, e := net.ResolveTCPAddr("tcp", a)
	if e != nil {
		log.Printf("error resolving tcp address: %v", e)
	}

	otherAddrs := make([]net.TCPAddr, 0)
	for i, a := range addresses {
		if i == server {
			continue
		}

		oAddr, e := net.ResolveTCPAddr("tcp", a)
		if e != nil {
			log.Printf("error resolving tcp address: %v", e)
		}

		otherAddrs = append(otherAddrs, *oAddr)
	}

	return *addr, otherAddrs
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var server int
	flag.IntVar(&server, "s", 0, "server number")
	flag.Parse()

	if server == -1 {
		log.Printf("server should be between 0 and 4")
		return
	}

	addr, others := splitAddresses(server)
	n := raft.InitNode(addr, others)
	e := n.Run()

	log.Printf("node stopped with error: %v", e)
}
