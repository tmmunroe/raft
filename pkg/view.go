package raft

import (
	"fmt"
	"net"
)

type RaftView struct {
	Epoch     int
	Leader    net.TCPAddr
	Followers []net.TCPAddr
}

func InitView(e int, l net.TCPAddr, f []net.TCPAddr) *RaftView {
	view := &RaftView{
		Epoch:     e,
		Leader:    l,
		Followers: make([]net.TCPAddr, 0),
	}
	view.Followers = append(view.Followers, f...)
	return view
}

func (rv RaftView) report() string {
	return fmt.Sprintf("view: epoch %v, leader %v, follower count %v", rv.Epoch, rv.Leader.String(), len(rv.Followers))
}

func (rv RaftView) Quorum() int {
	return (rv.Size() / 2) + 1
}

func (rv RaftView) Size() int {
	return 1 + len(rv.Followers)
}

func (rv RaftView) Same(rvOther RaftView) bool {
	return rv.Epoch == rvOther.Epoch &&
		SameAddress(rv.Leader, rvOther.Leader) &&
		SameAddresses(rv.Followers, rvOther.Followers)
}

func (rv RaftView) Clone() *RaftView {
	return &RaftView{
		Epoch:     rv.Epoch,
		Leader:    rv.Leader,
		Followers: rv.Followers,
	}
}

func (rv RaftView) IncrementView(newLeader net.TCPAddr) *RaftView {
	newFollowers := []net.TCPAddr{}
	for _, foll := range rv.Followers {
		if SameAddress(foll, newLeader) {
			continue
		}
		newFollowers = append(newFollowers, foll)
	}

	if !SameAddress(rv.Leader, newLeader) {
		newFollowers = append(newFollowers, rv.Leader)
	}

	return &RaftView{
		Epoch:     rv.Epoch + 1,
		Leader:    newLeader,
		Followers: newFollowers,
	}
}
