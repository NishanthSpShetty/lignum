# Lignum : Distributed Log store (akin kafka)

Lignum is a distributed message queue, implementing kafka in go using Consul as cluster management tool.

## Requirement
1. Consul service

You can run consule in docker with the following command
```
docker-compose up
```

## Test
```
make test
```

## Run
update the config.yml and run, make sure consul service is running and `consul` config are updated.
```
make run
```

To run on single host, change the server port for each instance.
