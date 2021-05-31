if [[ $# -ne 2 ]]; then
    echo "specify from and to in arguments"
    exit 
fi

from=$1
to=$2

 curl --request GET \
  --url http://localhost:8080/api/message \
  --header 'Content-Type: application/json' \
  --data "{ \"topic\":\"test\", \"from\":$1, \"to\": $2 }" | jq
