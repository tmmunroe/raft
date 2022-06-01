package raft

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type ConnectionOptions struct {
	Address *net.TCPAddr
}

func GetConnectionOptions(leaderIndex int) (*ConnectionOptions, error) {
	if leaderIndex < 0 || leaderIndex > 4 {
		return nil, fmt.Errorf("leaderIndex should be between 0 and 4")
	}

	leaderAddr, _ := SplitAddresses(leaderIndex)
	leader, err := net.ResolveTCPAddr(leaderAddr.Network(), leaderAddr.String())
	if err != nil {
		return nil, err
	}

	return &ConnectionOptions{Address: leader}, nil
}

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
