package raft

type RaftError string

const (
	NotLeader = RaftError("NotLeader")
)

func (re RaftError) Error() string {
	return string(re)
}
