# Lignum system configuration

## server config

> Root key : server

* server.host
 >
    Type      : Ip address
    Desription: Host address of the lignum service where it should bind itself. It can be localhsot or bind to 0.0.0.0

*  server.port
 >
    Type       : Number
    Value      : 8080
    Description: Port to which lignum should bind and listen on for incoming request.

* service-key
>
    Type       :  String
    Value      :  "service/lignum/key/master"
    Description:  A unique service key which is used for setting up a distributed lock for lignum cluster leader. Must be same for all node in a cluster.

----

## Cluster management config
> Root key: consul
* consul.host
 >
    Type      : Ip address
    Desription: Host address of the consul service.


* consul.port
 >
    Type      : Number
    Desription: Port of the consul service.

* consul.session-ttl-in-seconds

    Refer consul session [ttl](https://www.consul.io/api-docs/session#ttl)
 >
    Type      : Number
    Desription: consul session ttl.
    Value     : 15

* consul.session-renewal-ttl-in-seconds

  
 >
    Type      : Number
    Desription: consul session renewal ttl, client holding session must renew its session periodically before the `session-ttl-in-seconds` expires.
    Value     : 13
  
* consul.lock-delay-in-ms: 10

  Refer consul session [lockdelay](https://www.consul.io/api-docs/session#lockdelay)
  
>
    Type      : Number
    Desription: consul lock delay, when session is invalidated, it will not allow any other session to acquire repviously held lock fo specified time period.
    Value     : 15

* consul.leader-election-interval-in-ms
>
    Type      : Number
    Desription: Lignum periodically tries to elect itself as a leader, in case of leader is down and lock is released on the session key, it will acquire and elect itself as a leader, this key will control the frequency of attempt.
    Value     : 15

> for more refer detailed explanation in https://www.consul.io/docs/dynamic-app-config/sessions
----

## Message config

> Root key: message
*  message.initial-size-per-topic
>
    Type        : Number 
    Description : Topic buffer size, the number of message should be stored in a buffer, this also control number of messages store in wal and log files.

* message.data-dir
>
    Type        : String 
    Description : Data directory where messages should be persisted.
    Value       : "topics"
----

## Follower config

> Root key : follower
* follower.healthcheck-interval-in-seconds: 1
>
    Type        : Number 
    Description : Interval to check if all followers are healthy 
    Value       : 1

* follower.healthcheck-ping-timeout-in-ms
>
  
    Type        : Number 
    Description : Timeout for healthcheck on each follower
    Value       : 100


* follower.registration-leader-check-interval-in-seconds
>
    Type       : Number
    Desription : Interval to check if there is a change in cluster leader by querying consul, if changed connect to the new leader as follower.


----

## Replication config

> Root key : replication
* replication.internal-queue-size
>
    Type       : Number 
    Desription : Size of internal buffer/channel which holds messages to be sent over to follower 
    Value      : 1024
    
* replication.client-timeout-in-ms
>
    Type        : Number 
    Description : Timeout for replication write on each follower
    Value       : 25


Refer the default config [here](../config.yml)