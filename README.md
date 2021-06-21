## Lignum : Distributed Log store (akin kafka)

Lignum is a distributed message queue, implementing kafka in go using Consul as cluster management tool.

![Status](https://github.com/NishanthSpShetty/lignum/actions/workflows/go.yml/badge.svg)
[![NishanthSpShetty](https://circleci.com/gh/NishanthSpShetty/lignum.svg?style=svg&circle-token=4de78d34762f2fe9f94fdbfc8cb5d29b146e813b)](https://app.circleci.com/pipelines/github/NishanthSpShetty/lignum)
[![CodeFactor](https://www.codefactor.io/repository/github/nishanthspshetty/lignum/badge?s=82e5d72d47892bd920b35d26664d7d3b0643cdd8)](https://www.codefactor.io/repository/github/nishanthspshetty/lignum)

### Functionality
Simple message queue inspired by Kafka, which can be used to
   * send messages to topic.
   * consume messages from topic.

### Cluster
* Lignum can operate in cluster mode without needing any special setup.
* Cluster management is facilitated by consul.
* Each node will connect to consul to get the leader, if no leader found one of the node will be elected as leader.
* All other node will register itself as follower to the leader.
* Message sent to leader, will be replicated to follower node [WIP]



Lignum message has the following characterstics

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

> NOTE: Lignum doesnt have API to create new topic, it will create a topic if doesnt exist.

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
> NOTE: Lignum doesnt store any data about the consumer. 
---
