#!/bin/bash
echo "Generating messages "

count=0

while read -r msg
do 
    curl --request POST \
        --url http://localhost:8080/api/message \
        --header 'Content-Type: application/json' \
        --data "{ \"topic\": \"dmesg-log\", \"message\":\"${msg}\" }" | jq;
            (( count++ ))
        done < <(dmesg)
        echo "Sent $count messages"


