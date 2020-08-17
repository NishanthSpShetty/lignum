package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
)

const (
	//Register this key on consul and get  a lock on it.
	ServiceKey  = "service/disqlogger/key/master"
	ServiceName = "distributed-quote-logger"
	ttlS        = "10s"
)

var isLeader = false
var serviceId uuid.UUID
var client *api.Client
var HOST string
var PORT int

func signalHandler(sessionRenewalChannel chan struct{}, sessionId string) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	//read the signal and discard
	<-signalChannel
	log.Infof("Destroying session and stopping service [ServiceID : %s ]", serviceId.String())
	close(sessionRenewalChannel)

	_, err := client.Session().Destroy(sessionId, nil)

	if err != nil {
		log.Error("Failed to destroy the session ", err)
	}
	os.Exit(0)
}

//connectToLeader Connect thsi service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func connectToLeader() {

	requestBody, _ := json.Marshal(map[string]interface{}{
		"node": serviceId,
		"host": HOST,
		"PORT": PORT,
	})
	for {
		//loop if the current node becomes the leader
		log.Infoln("Registering this servce as a follower to the clusetr leader...")
		//get the leader
		kv, _, err := client.KV().Get(ServiceKey, nil)
		if err != nil {
			log.Errorln("Failed to get the leader information...")
			//TODO: give it a second and loop back??
			time.Sleep(1 * time.Second)
			continue
		}
		//check if the KV has a session
		if !isLeader {
			if kv != nil && kv.Session != "" {
				//get the leader information and send a follow request.
				leaderEndpoint := fmt.Sprintf("http://localhost:%s%s", kv.Value, "/service/api/follower/register")
				resp, err := http.Post(leaderEndpoint, "application/json", bytes.NewBuffer(requestBody))
				if err != nil {
					log.Errorln("Failed to register with the leader ", err)
					return
				}
				response, err := ioutil.ReadAll(resp.Body)
				log.Infof("ConnecToLeader Response : %s\n ", string(response))
				break
			}
			//send connect ping to leader
		} else {
			log.Infoln("Im the leader....")
			return
		}
	}
}

//leaderElection Function will keep trying to aquire lock on the `ServiceKey`
func leaderElection(port int, sessionId string) {
	//make sure there is a leader registered on the consul all time.
	loggedOnce := false
	for {
		if !isLeader {

			aquired, queryDuration, err := client.KV().Acquire(&api.KVPair{
				Key:     ServiceKey,
				Value:   []byte(fmt.Sprintf("%d", port)),
				Session: sessionId,
			}, nil)

			if err != nil {
				log.Errorf("Failed to aquire lock %v \n", err)
				continue
			}

			if aquired {
				isLeader = aquired
				log.Infof("Lock aquired and marking the service as leader, Lock aquired in %dms\n", queryDuration.RequestTime.Milliseconds())
			} else {
				if !loggedOnce {
					log.Debug("Lock is already taken, will check again...")
					loggedOnce = true
				}
			}

			time.Sleep(60 * time.Millisecond)

		} else {
			//if once the leader elected, stop the busy loop, will change this later if need to reaquire the lock
			return
		}
	}
}

func main() {

	_, err := net.InterfaceAddrs()
	HOST = "localhost" //addr[0].String()

	//parse the command line args
	flag.IntVar(&PORT, "port", 9090, "Service port")
	flag.Parse()

	serviceId = uuid.New()

	log.SetLevel(log.DebugLevel)

	log.Info("Starting loger service [ServiceID : %s ].", serviceId.String())
	config := api.DefaultConfig()

	config.Address = "localhost:8500"
	client, err = api.NewClient(config)

	if err != nil {
		log.Error("Failed to create the new consul client ", err)
		return
	}

	sessionEntry := &api.SessionEntry{
		Name:      ServiceName,
		TTL:       ttlS,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, queryDuration, err := client.Session().Create(sessionEntry, nil)

	if err != nil {
		log.Errorf("Failed to create session %v \n", err)
		return
	}

	log.Debugf("Consul session created, ID : %v, Aquired in :%dms  ", sessionId, queryDuration.RequestTime.Milliseconds())

	//Start leader election routine
	go leaderElection(PORT, sessionId)
	doneChan := make(chan struct{})
	go signalHandler(doneChan, sessionId)
	//connect to leader

	//TODO: try to connect to the leader, if not found call the leader election routine to make this service as the leader,
	//so we shouyld start the leader connection routine.
	go connectToLeader()
	//Create a session renewer routine.
	go func() {
		defer close(doneChan)
		for {
			client.Session().RenewPeriodic(ttlS, sessionId, nil, doneChan)
			time.Sleep(90 * time.Second)
		}
	}()

	log.Infof("Starting HTTP service at %s:%d \n", HOST, PORT)

	//start the work.
	address := fmt.Sprintf("%s:%d", HOST, PORT)
	http.HandleFunc("/service/api/follower/register", func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Infof("Request recieved for follower registration %v ", string(requestBody))

		fmt.Fprintf(w, "Follower registered service  %s\n", serviceId)
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "PONG")
	})
	log.Panic(http.ListenAndServe(address, nil))
}
