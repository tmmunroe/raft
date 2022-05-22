package raft

import "net"

type RaftView struct {
	Epoch     int
	Leader    *net.TCPAddr
	Followers []*net.TCPAddr
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

func (rv RaftView) IncrementView(newLeader *net.TCPAddr) *RaftView {
	if SameAddress(rv.Leader, newLeader) {
		return nil
	}

	newFollowers := []*net.TCPAddr{rv.Leader}
	for _, foll := range rv.Followers {
		if SameAddress(foll, newLeader) {
			continue
		}
		newFollowers = append(newFollowers, foll)
	}

	return &RaftView{
		Epoch:     rv.Epoch + 1,
		Leader:    newLeader,
		Followers: newFollowers,
	}
}
