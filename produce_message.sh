#!/bin/bash
echo "Generating messages "

dmesg | while read msg
do 
    echo  "{ \"Message\":\"${msg}\" }";
    curl --request POST \
        --url http://localhost:8080/api/message \
        --header 'Content-Type: application/json' \
        --data "{ \"Message\":\"${msg}\" }" ;
    done
