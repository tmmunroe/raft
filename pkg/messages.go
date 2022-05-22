package raft

type AppendArgs struct {
	View    RaftView
	Entries []LogEntry
}

type AppendReply struct {
	Accepted bool
	View     RaftView
}

type RequestVoteArgs struct {
	Proposed RaftView
}

type RequestVoteReply struct {
	Accepted bool
}
