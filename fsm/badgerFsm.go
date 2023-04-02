package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"io"
	"os"
	"strings"
)

type badgerFSM struct {
	db *badger.DB
}

func (b badgerFSM) get(key string) (interface{}, error) {
	var data interface{}

	txn := b.db.NewTransaction(false)

	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get([]byte(key))
	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	var value = make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	if value != nil && len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		data = map[string]interface{}{}
	}
	return data, err
}

type ValueWithLogIndex struct {
	Value    string
	LogIndex raft.Log
}

// key-value change to key-valueWithLogIndex
func (b badgerFSM) set(key string, value interface{}) error {
	var data = make([]byte, 0)

	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	if data == nil || len(data) <= 0 {
		return nil
	}

	txn := b.db.NewTransaction(true)
	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return err
	}

	return txn.Commit()
}

func (b badgerFSM) delete(key string) error {
	txn := b.db.NewTransaction(true)
	err := txn.Delete([]byte(key))

	if err != nil {
		txn.Discard()
		return err
	}

	return txn.Commit()
}

func (b badgerFSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var payload = CommandPayload{}
		if err := json.Unmarshal(log.Data, &payload); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error marshalling store payload %s\n", err.Error())
			return nil
		}

		op := strings.ToUpper(strings.TrimSpace(payload.Operation))
		switch op {
		case "SET":
			return &ApplyResponse{
				Error: b.set(payload.Key, payload.Value),
				Data:  payload.Value,
			}
		case "GET":
			data, err := b.get(payload.Key)
			return &ApplyResponse{
				Error: err,
				Data:  data,
			}
		case "DELETE":
			err := b.delete(payload.Key)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil
}

func (b badgerFSM) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshotNoop()
}

func (b badgerFSM) Restore(rClose io.ReadCloser) error {
	defer func() {
		if err := rClose.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[FINALLY RESTORE] close error %s\n", err.Error())
		}
	}()

	_, _ = fmt.Fprintf(os.Stdout, "[START RESTORE] read all message from snapshot\n")

	var totalRestored int
	decoder := json.NewDecoder(rClose)
	for decoder.More() {
		var data = &CommandPayload{}
		err := decoder.Decode(data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error decode data %s\n", err.Error())
			return err
		}

		if err := b.set(data.Key, data.Value); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error persist data %s\n", err.Error())
			return err
		}

		totalRestored++
	}

	_, err := decoder.Token()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error %s\n", err.Error())
	}

	_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] success restore %d messages in snapshot\n", totalRestored)
	return nil
}

// NewBadger raft.FSM implementation using badgerDB
func NewBadger(badgerDB *badger.DB) raft.FSM {
	return &badgerFSM{
		db: badgerDB,
	}
}
