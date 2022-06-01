package main

import (
	"flag"
	"fmt"
	"log"

	raft "github.com/tmmunroe/raft/pkg"
)

func main() {
	var leaderIndex int
	flag.IntVar(&leaderIndex, "s", 0, "leader number")
	flag.Parse()

	opts, err := raft.GetConnectionOptions(leaderIndex)
	if err != nil {
		log.Printf("error getting connection options: %v", err)
	}

	client := raft.InitRaftClient()
	err = client.Connect(*opts)
	if err != nil {
		log.Printf("error connecting to leader: %v", err)
	}

	client.Put("a", "a old value")
	fmt.Print(client.Get("a"))

	client.Put("b", "b old value")
	fmt.Print(client.Get("b"))

	client.Put("a", "a new value")
	fmt.Print(client.Get("a"))

	client.Put("b", "b new value")
	fmt.Print(client.Get("b"))
}
