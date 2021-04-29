package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lignum/cluster"
	"github.com/lignum/config"
	"github.com/lignum/message"
	log "github.com/sirupsen/logrus"
)

func registerFollower(serviceId string) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Infof("Request received for follower registration %v ", string(requestBody))
		node := cluster.Node{}
		json.Unmarshal(requestBody, &node)
		cluster.AddFollower(node)
		fmt.Fprintf(w, "Follower registered. Node : [ %v ]\n", node)
		fmt.Printf(" Current followers \n %v ", cluster.GetFollowers())
	}
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "PONG")
}

// request message struct
type PutMessageRequest struct {
	Message string
}

type GetMessageRequest struct {
	//will need range to pick the messages from
	From int
	To   int
}

//respons message struct
type GetMessageResponse struct {
	Messages []string `json:"messages"`
}

//handleMessagePut Write the message with the given key.
func handleMessage(messageChannel chan<- message.MessageT) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()

		switch req.Method {
		case "POST":
			var messageRequest PutMessageRequest
			err := decoder.Decode(&messageRequest)
			if err != nil {
				log.Errorf("Failed to read request body %s ", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Debugf("Recieved message %v \n", messageRequest)
			message.Put(messageRequest.Message)

			fmt.Fprintf(w, "{status : \"message commited\"\n message : { %v }", "key:value")

		case "GET":
			var messageRequest GetMessageRequest = GetMessageRequest{}

			err := decoder.Decode(&messageRequest)

			if err != nil {
				log.Errorf("Failed to read request body %s ", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			messages := message.Get(messageRequest.From, messageRequest.To)
			messag := GetMessageResponse{Messages: messages}

			log.Debugf(" Recieved message %v ", messageRequest)
			json.NewEncoder(w).Encode(messag)
		default:
			http.Error(w, "request method must be one of [ GET, POST ].", http.StatusMethodNotAllowed)
		}
	}
}

func StartApiService(appConfig config.Config, serviceId string, messageChannel chan<- message.MessageT) {

	log.Infof("Starting HTTP service at %s:%d \n", appConfig.Server.Host, appConfig.Server.Port)
	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/service/api/follower/register", registerFollower(serviceId))
	http.HandleFunc("/api/message", handleMessage(messageChannel))
	log.Panic(http.ListenAndServe(address, nil))
}
