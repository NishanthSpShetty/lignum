package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/rs/zerolog/log"
)

func registerFollower(serviceId string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Info().Bytes("RequestBody", requestBody).Msg("Request received for follower registration")
		node := cluster.Node{}
		json.Unmarshal(requestBody, &node)
		cluster.AddFollower(node)
		fmt.Fprintf(w, "Follower registered. Node : [ %v ]\n", node)
		//		fmt.Printf(" Current followers \n %v ", cluster.GetFollowers())
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
				log.Error().Err(err).Msg("Failed to read request body %s ")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Debug().Interface("RecievedMessage", messageRequest).Send()
			message.Put(messageRequest.Message)

			fmt.Fprintf(w, "{status : \"message commited\"\n message : { %v }", "key:value")

		case "GET":
			var messageRequest GetMessageRequest = GetMessageRequest{}

			err := decoder.Decode(&messageRequest)

			if err != nil {
				log.Error().Err(err).Msg("Failed to read request body")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			messages := message.Get(messageRequest.From, messageRequest.To)
			messag := GetMessageResponse{Messages: messages}

			log.Debug().Interface("RecievedMessage", messageRequest).Send()
			json.NewEncoder(w).Encode(messag)
		default:
			http.Error(w, "request method must be one of [ GET, POST ].", http.StatusMethodNotAllowed)
		}
	}
}

func StartApiService(appConfig config.Config, serviceId string, messageChannel chan<- message.MessageT) error {

	log.Info().
		Str("Host", appConfig.Server.Host).
		Int("Port", appConfig.Server.Port).
		Msg("Starting HTTP service")

	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/service/api/follower/register", registerFollower(serviceId))
	http.HandleFunc("/api/message", handleMessage(messageChannel))
	return http.ListenAndServe(address, nil)
}
