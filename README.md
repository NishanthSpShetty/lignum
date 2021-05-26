## Lignum : Distributed Log store (akin kafka)

Lignum is a distributed message queue, implementing kafka in go using Consul as cluster management tool.

![Status](https://github.com/NishanthSpShetty/lignum/actions/workflows/go.yml/badge.svg)
[![NishanthSpShetty](https://circleci.com/gh/NishanthSpShetty/lignum.svg?style=svg&circle-token=4de78d34762f2fe9f94fdbfc8cb5d29b146e813b)](https://app.circleci.com/pipelines/github/NishanthSpShetty/lignum)
[![CodeFactor](https://www.codefactor.io/repository/github/nishanthspshetty/lignum/badge?s=82e5d72d47892bd920b35d26664d7d3b0643cdd8)](https://www.codefactor.io/repository/github/nishanthspshetty/lignum)

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

To run on single host, change the server port for each instance.
