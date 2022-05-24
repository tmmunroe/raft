package raft

import "net/rpc"

type RaftClient struct {
	c *rpc.Client
}
