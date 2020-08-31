consul in docker container  ==> https://learn.hashicorp.com/consul/day-0/containers-guide

run consul server
docker run -d -p 8500:8500 -p 8600:8600/udp --name=badger consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

## Distribute Quote Logger (should be renamed)

Create a disributed logging service using golang, something like go.

## TODO's

### cluster 
 - [X] Leader election
 - [] Get the current cluster leader and register to them.
 - [X] Elect any node as new leader if current leader fais fort some reason.
 - [] When new leader is elected, all follower should register themselves with the leader again.

### functional
  - [] Send message to any node in a cluster which should be replicated to all node.
  - [] Can read from any node.




