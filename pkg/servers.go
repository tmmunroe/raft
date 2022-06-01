package raft

import (
	"log"
	"net"
)

var Addresses = []string{
	"127.0.0.1:51001",
	"127.0.0.1:51003",
	"127.0.0.1:51005",
	"127.0.0.1:51007",
	"127.0.0.1:51009",
}

func SplitAddresses(server int) (net.TCPAddr, []net.TCPAddr) {
	a := Addresses[server]

	addr, e := net.ResolveTCPAddr("tcp", a)
	if e != nil {
		log.Printf("error resolving tcp address: %v", e)
	}

	otherAddrs := make([]net.TCPAddr, 0)
	for i, a := range Addresses {
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
