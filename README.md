# Distributed Key-Value Store
- Use Raft as consensus module
- Use BotlDB as log store
- Use BadgerDB as persistent storage

# Commands
## To start a node with http address and raft rpc address
```SERVER_PORT=<SERVER_PORT> RAFT_NODE_ID=<NODE_ID> RAFT_PORT=<RAFT_PORT> RAFT_VOL_DIR=<RAFT_VOLUME_DIR> run main.go```

# Follower Read (quorum read)
## Cannot perform 
Although a replica can query all nodes' addresses of the current cluster, 
these addresses are used only for AppendEntries and RequestVote RPCs

## Two ways to achieve this 
- add another rpc/http address that Raft needs to keep track 
and allows custom handlers for this address (need to modify Hashicorp/Raft package)
- client needs to perform quorum reads on non-leader replica

