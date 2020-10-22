package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

func StartApiService(appConfig config.Config, serviceId string) {

	log.Infof("Starting HTTP service at %s:%d \n", appConfig.Server.Host, appConfig.Server.Port)
	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/service/api/follower/register", func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Infof("Request received for follower registration %v ", string(requestBody))

		fmt.Fprintf(w, "Follower registered. Service ID : [ %s ]\n", serviceId)
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "PONG")
	})
	log.Panic(http.ListenAndServe(address, nil))
}
