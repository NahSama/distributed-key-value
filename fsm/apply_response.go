package fsm

// ApplyResponse from Apply raft
type ApplyResponse struct {
	Error error
	Data  interface{}
}
