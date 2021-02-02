package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

func registerFollower(serviceId string) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)

		log.Infof("Request received for follower registration %v ", string(requestBody))
		fmt.Fprintf(w, "Follower registered. Service ID : [ %s ]\n", serviceId)
	}
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "PONG")
}

// request message struct
type PutMessageRequest struct {
	Key   string
	Value string
}

type GetMessageRequest struct {
	Key string
}

//respons message struct
type PutMessageResponse struct {
	Key   string
	Value string
}

//handleMessagePut Write the message with the given key.
func handleMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		switch req.Method {
		case "POST":
			var messageRequest PutMessageRequest
			err := json.NewDecoder(req.Body).Decode(&messageRequest)
			if err != nil {
				log.Infof("Failed to read request body %s ", err)
			}
			log.Debugf("Recieved message %v \n", messageRequest)
			fmt.Fprintf(w, "{status : \"message commited\"\n message : { %v }", "key:value")
		case "GET":
			var messageRequest GetMessageRequest

			err := json.NewDecoder(req.Body).Decode(&messageRequest)
			if err != nil {
				log.Infof("Failed to read request body %s ", err)
			}

			log.Debugf(" Recieved message %v ", messageRequest)
			fmt.Fprintf(w, "{ message : { %v }", "key:value")

		default:
			http.Error(w, "request method must be one of [ GET, POST ].", http.StatusMethodNotAllowed)
		}
	}
}

func StartApiService(appConfig config.Config, serviceId string) {

	log.Infof("Starting HTTP service at %s:%d \n", appConfig.Server.Host, appConfig.Server.Port)
	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/service/api/follower/register", registerFollower(serviceId))
	http.HandleFunc("/api/message", handleMessage())
	log.Panic(http.ListenAndServe(address, nil))
}
