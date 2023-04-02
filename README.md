#Follower Read (quorum read)
## Cannot perform 
Even replica can query all nodes' addresses of the current cluster, 
these addresses are used for only AppendEntries and RequestVote RPCs

## Two ways to achieve this 
- add another rpc/http address that Raft needs to keep track 
and allows custom handlers for this address (need to modify Hashicorp/Raft package)
- client needs to perform quorum reads on non-leader replica

