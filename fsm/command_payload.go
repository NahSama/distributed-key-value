package fsm

// CommandPayload is payload send by calling raft.Apply(cmd []byte, timeout time.Duration)
//type ValueWithTimestamp struct {
//	Timestamp time.Time
//	Value     interface{}
//}

type CommandPayload struct {
	Operation string
	Key       string
	Value     interface{}
}
