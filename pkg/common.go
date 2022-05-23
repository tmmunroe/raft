package raft

import (
	"math/rand"
	"net"
	"time"
)

func RandomDuration(min int, max int) time.Duration {
	randInt := min + rand.Intn(max-min)
	return time.Duration(randInt) * time.Second
}

func SameAddress(a net.TCPAddr, b net.TCPAddr) bool {
	return a.Network() == b.Network() &&
		a.String() == b.String()
}

func SameAddresses(fa []net.TCPAddr, fb []net.TCPAddr) bool {
	if len(fa) != len(fb) {
		return false
	}

	for i := range fa {
		if !SameAddress(fa[i], fb[i]) {
			return false
		}
	}

	return true
}
