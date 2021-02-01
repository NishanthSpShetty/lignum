package api

import (
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

//handleMessagePut Write the message with the given key.
func handleMessagePut() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			requestBody, _ := ioutil.ReadAll(req.Body)
			log.Infof("Write log %s ", requestBody)
		}
	}
}

//handleMessageGet Get the message for the given key
func handleMessageGet() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			requestBody, _ := ioutil.ReadAll(req.Body)
			log.Infof("Get log %s ", requestBody)
		}
	}
}

func StartApiService(appConfig config.Config, serviceId string) {

	log.Infof("Starting HTTP service at %s:%d \n", appConfig.Server.Host, appConfig.Server.Port)
	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/service/api/follower/register", registerFollower(serviceId))
	http.HandleFunc("api/message/put", handleMessagePut())
	http.HandleFunc("api/message/get", handleMessageGet())
	log.Panic(http.ListenAndServe(address, nil))
}
