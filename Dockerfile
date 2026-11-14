FROM docker.io/library/golang:1.27.1-alpine as builder

# MAINTAINER Nishanth Shetty nishanthspshetty@gmail.com

WORKDIR /app
COPY . .

RUN go get -v ./... && go build -o ./lignum


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/lignum . 
EXPOSE 8080
CMD  [ "./lignum", "-config", "config.yml" ]

