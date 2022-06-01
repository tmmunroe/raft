package raft

type AppendArgs struct {
	View           RaftView
	LastIndex      int
	Entries        []LogEntry
	CommittedIndex int
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

type ClientCommandArgs struct {
	Id      int64
	Command string
	Args    MapCommandArgs
}

type ClientCommandReply struct {
	Id      int64
	Command string
	Result  MapCommandResult

	Error error
	View  RaftView
}
