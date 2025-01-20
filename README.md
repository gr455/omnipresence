# Omnipresence
Replicated key-value storage. etcd's cousin.

Omnipresence uses the Raft consensus algorithm for its replication engine.

## Run
1. Run `make`
2. Within the `bin/` directory, find the `omnipresence` binary
3. Set `RAFT_PEER_ID` and `RAFT_PEER_PORT` env vars and run the binary to start a single Raft node.
```
RAFT_PEER_ID=peer1 RAFT_PEER_PORT=50051 bin/linux-amd64/omnipresence
```
4. Start many Raft nodes
5. Within the `bin/` directory, find the `client` binary
6. Run the client binary to interact with the key value map.

![Screenshot](https://github.com/user-attachments/assets/756e19cc-63e6-481a-84fc-fe72c09e7aef)


