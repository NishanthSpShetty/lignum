FROM ubuntu:18.04
MAINTAINER Nishanth Shetty nishanthspshetty@gmail.com

WORKDIR .
COPY ./bin/lignum .
COPY ./config.yml .

CMD  [ "./lignum", "-config", "config.yml" ]
