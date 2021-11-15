## Lignum : Distributed Log store (akin kafka)

Lignum is a distributed message queue, implementing something like kafka in go using Consul as cluster management service.

![Status](https://github.com/NishanthSpShetty/lignum/actions/workflows/go.yml/badge.svg)
[![NishanthSpShetty](https://circleci.com/gh/NishanthSpShetty/lignum.svg?style=svg&circle-token=4de78d34762f2fe9f94fdbfc8cb5d29b146e813b)](https://app.circleci.com/pipelines/github/NishanthSpShetty/lignum)
[![CodeFactor](https://www.codefactor.io/repository/github/nishanthspshetty/lignum/badge?s=82e5d72d47892bd920b35d26664d7d3b0643cdd8)](https://www.codefactor.io/repository/github/nishanthspshetty/lignum)

### Motivation
I set out myself to learn distributed system, while doing so started with distributed lock using consul and expanded the project to build logger and eventually decided to build something like kafka. So here I am now.

### Architecture diagram
![Lignum cluster](https://github.com/NishanthSpShetty/lignum/blob/master/resources/diagrams/lignum_system_design.png)


### Functionality
Simple message queue inspired by Kafka, which can be used to
   * Send messages to topic.
   * Consume messages from topic.
   * Message is replicated to all nodes. meaning any node can be used to read message from the topic. compared to kafka where topic has leader and replicated to few set of nodes in a cluster.
   * Persist messages on disk.
   
   
### Cluster
* Lignum can operate in cluster mode without needing any special setup.
* Cluster management is facilitated by consul.
* Each node will connect to consul to get the leader, if no leader found one of the node will be elected as leader.
* All other node will register itself as follower to the leader.
* Message sent to leader, will be replicated to follower node.
* If the leader goes down, any one of the follower gets elected as a leader.


#### Lignum message has the following characterstics

1. Topic

    Each message belongs to its own topic, so that way you can use lignum to send  messages from multiple services/use case and consume the same later from the topic

2.  Message

    The actual message data which we want to write to given topic

---

### Requirement
Consul service

You can run consule in docker with the following command
```
docker-compose up
```

### Test
```
make test
```

### Run
update the config.yml and run, make sure consul service is running and `consul` config are updated.
```
make run
```

For development, lignum can log message as simple readable text on the console, set the environemnt variable ENV to development 
```
export ENV="development"
```

To set the log level, use
```
export LOG_LEVEL="error"
```

To create a cluster on single host, change the server port for each instance.
Lignum will listen on the specified port for incoming traffics

---


## Configuration
 
 Refer configuration [document](./doc/config.md)

---

### Usage

#### Sending message

```
Endpoint    /api/message

Method     POST

Request   {"topic"  : "beautiful_topic_name", 
           "message": "message from oracle"}
```

Example curl
```
curl --request POST \
  --url http://localhost:8080/api/message \
  --header 'Content-Type: application/json' \
  --data '{
	"topic": "test",
	"message":"this is a test message"
}'
```

For this server will respond with the following
```
{
  "status": "message commited",
  "data": "this is a test message"
}
```

> NOTE: Lignum doesnt have API to create new topic as of now, it will create a topic if doesnt exist.

---
#### Read message from lignum

```
Endpoint    /api/message

Method     GET

Request   {"topic"  : "beautiful_topic_name", 
           "from": 0,
           "to": 100}
```

where

    topic: topic you wish to consume message from
    from : message offset you are reading from,
    to   : message offset upto the given `to` value. (excluding)


Example curl

```
curl --request GET \
  --url http://localhost:8080/api/message \
  --header 'Content-Type: application/json' \
  --data '{
	"topic": "test",
	"from": 0,
	"to": 3
}'
```

the above message will return 3 messages if presents, if the message is less than what `to` offset specified, it will return all messages in topic.

```
{
  "messages": [
    {
      "Id": 0,
      "Data": "this is a test message"
    },
    {
      "Id": 1,
      "Data": "this is a test message 2"
    },
    {
      "Id": 2,
      "Data": "this is a test message 3"
    }
  ],
  "count": 3
}
```

>> NOTE: Lignum doesnt store any data about the consumer, so it wont track the last message consumed as done by kafka. 

## Storage

Lignum stores the messages to disk as lomg files, batched with the messageBufferSize  as specified with config "initial-size-per-topic".

It creates a WAL file and when the umber of message reaches set limit with buffer size, it will flused to log file and new wal file is created with the different offset for message.




---

*For TODO's and progress on the project refer [laundrylist](https://github.com/NishanthSpShetty/lignum/blob/master/laundrylist.md) or [lignum project](https://github.com/NishanthSpShetty/lignum/projects/1)*

---

