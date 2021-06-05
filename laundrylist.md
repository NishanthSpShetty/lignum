consul in docker container  ==> https://learn.hashicorp.com/consul/day-0/containers-guide

run consul server
docker run -d -p 8500:8500 -p 8600:8600/udp --name=badger consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

## Lignum : Distributed message queue

Create a disributed logging/messaging service using golang, something like kafka.

## TODO's

### Good to have
 - [X] Read config from the config file.
 - [] Move to docker multiple node cluster setup. -WIP
 - [X] implement http api 
 - [] Implement core functionality
 - [X] Unit test for cluster functions
 
### cluster 
 - [X] Leader election
 - [X] Get the current cluster leader and register to them.
 - [] Health check of all the followers, if its not available remove them from the list of followers
 - [X] Elect any node as new leader if current leader fails for some reason.
 - [X] When new leader is elected, all follower should register themselves with the leader again.
 - [X] Each node must know the leader.

### functional
  - [X] Send message and read from the topic
  - [X] Create topic if doesnt not exist
  - [X] Each topic has a separate offset conter
  - [] send the message to leader for replication
  - [] Which should be replicated to all node by leader.
  - [] Can read from any node.

