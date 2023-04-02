package store_handler

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
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
