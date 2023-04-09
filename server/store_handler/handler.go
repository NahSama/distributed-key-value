package store_handler

import (
	raft "github.com/NahSama/raft-modified"
	"github.com/dgraph-io/badger/v2"
)

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Response struct {
	Data  KeyValue `json:"data"`
	Error error    `json:"error"`
}

type handler struct {
	raft *raft.Raft
	db   *badger.DB
}

func New(raft *raft.Raft, db *badger.DB) *handler {
	return &handler{
		raft: raft,
		db:   db,
	}
}
