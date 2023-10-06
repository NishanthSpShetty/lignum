### Note
[consul in docker container](https://learn.hashicorp.com/consul/day-0/containers-guide)

### run consul server
```
docker run -d -p 8500:8500 -p 8600:8600/udp --name=badger consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

## Lignum : Distributed message queue

Create a disributed logging/messaging service using golang, something like kafka.


### Good to have
 - [X] Read config from the config file.
 - [ ] Move to docker multiple node cluster setup. -WIP
 - [X] implement http api.
 - [X] Unit test for cluster functions.
 - [X] Script to produce and consume messages.
 - [X] gRPC API
 - [ ] Topic metadata persistance, currently topic metadata is read from data segment stored in wal file.


### cluster 
 - [X] Leader election
 - [X] Get the current cluster leader and register to them.
 - [X] Health check of all the followers, if its not available mark them as unhealthy or remove them from the follower registry.
 - [X] Elect any node as new leader if current leader fails for some reason.
 - [X] When new leader is elected, all follower should register themselves with the leader again.
 - [X] Each node must know the leader.

### functional
  - [X] Send message to and read from the topic.
  - [X] Create topic if doesnt not exist when the first message is published.
  - [X] Each topic has a separate offset conter.
  - [ ] send the message to leader for replication in case follower receives the message. [not in scope]
  - [X] When message received by leader, it should be replicated to all healthy follower node.
  - [X] Can read from any node.
  - [X] Message is persisted by each node.
  - [X] WAL replication
  - [ ] Switching between live and WAL replication
  - [ ] Server metrics, currently has fewer metrics published in `/metrics` api, need to add more metrics on various points.
