package raft_handler

import (
	raft "github.com/NahSama/raft-modified"
)

type handler struct {
	raft *raft.Raft
}

func New(raft *raft.Raft) *handler {
	return &handler{
		raft: raft,
	}
}
