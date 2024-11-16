package main

import (
	"fmt"
	"github.com/gr455/omnipresence/raft"
	"github.com/gr455/omnipresence/raft/storage"
)

func main() {

	s := storage.NewRaftStorage("./raft/storage/log.txt", "./raft/storage/persistent.txt")

	r, err := raft.NewRaftConsensusObject("1", s, 1)
	if err != nil {
		return
	}

	fmt.Println(r)

	return
}
