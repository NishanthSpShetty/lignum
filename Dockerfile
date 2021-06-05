FROM golang:1.16
# MAINTAINER Nishanth Shetty nishanthspshetty@gmail.com

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./... && go build -o ./lignum
CMD  [ "./lignum", "-config", "config.yml" ]

EXPOSE 8080
