package raft

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type RaftClient struct {
	client *rpc.Client
	leader *net.TCPAddr
	epoch  int64
	mu     sync.Mutex
}

func InitRaftClient() *RaftClient {
	return &RaftClient{
		client: nil,
		leader: nil,
		epoch:  1,
		mu:     sync.Mutex{},
	}
}

func (rc *RaftClient) nextId() int64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.epoch += 1
	return rc.epoch
}

func (rc *RaftClient) Connect(opts ConnectionOptions) error {
	log.Printf("connecting to leader at %v...", opts.Address.String())
	leader, err := net.ResolveTCPAddr(opts.Address.Network(), opts.Address.String())
	if err != nil {
		return err
	}

	client, err := rpc.Dial(leader.Network(), leader.String())
	if err != nil {
		return err
	}

	log.Printf("connected...")
	rc.leader = leader
	rc.client = client
	return nil
}

func (rc *RaftClient) Put(key string, value string) error {
	if rc.client == nil {
		return fmt.Errorf("client not connected")
	}

	args := &ClientCommandArgs{
		Id:      rc.nextId(),
		Command: PutValue,
		Args:    MapCommandArgs{Key: key, Value: value},
	}
	reply := &ClientCommandReply{}

	log.Printf("Put args: %+v", args)
	err := rc.client.Call("ClientService.Request", args, reply)
	if err != nil {
		return err
	}
	log.Printf("Put reply: %+v", reply)

	if reply.Error != nil {
		if reply.Error.Error() == NotLeader.Error() {
			leader, _ := net.ResolveTCPAddr(reply.View.Leader.Network(), reply.View.Leader.String())
			rc.leader = leader
			return reply.Error
		}
	}

	return nil
}

func (rc *RaftClient) Get(key string) (string, error) {
	if rc.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	args := &ClientCommandArgs{
		Id:      rc.nextId(),
		Command: GetValue,
		Args:    MapCommandArgs{Key: key},
	}
	reply := &ClientCommandReply{}

	log.Printf("Get args: %+v", args)
	err := rc.client.Call("ClientService.Request", args, reply)
	if err != nil {
		return "", err
	}
	log.Printf("Get reply: %+v", reply)

	if reply.Error != nil {
		if reply.Error.Error() == NotLeader.Error() {
			leader, _ := net.ResolveTCPAddr(reply.View.Leader.Network(), reply.View.Leader.String())
			rc.leader = leader
			return "", reply.Error
		}
	}

	return reply.Result.Value, nil
}
