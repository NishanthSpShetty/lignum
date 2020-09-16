package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/disquote-logger/config"
	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

//isLeader mark it as true when the current node becomes the leader
var isLeader = false

//connectToLeader Connect thsi service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func ConnectToLeader(appConfig config.Server, serviceId string) {

	requestBody, _ := json.Marshal(map[string]interface{}{
		"node": serviceId,
		"host": appConfig.Host,
		"PORT": appConfig.Port,
	})
	for {
		//loop if the current node becomes the leader
		log.Infoln("Registering this servce as a follower to the clusetr leader...")
		//get the leader
		kv, err := GetLeader(appConfig.ServiceKey)
		if err != nil {
			log.Errorf("Failed to get the leader information", err)
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
			} else {
				log.Errorln("Unable to get the leader information, will try again")
			}
			//send connect ping to leader
		} else {
			log.Infoln("Im the leader....")
			return
		}
	}
}
func leaderElection(nodeConfig NodeConfig, sessionId string, serviceKey string) {
	loggedOnce := false
	kvPair := &api.KVPair{
		Key:     serviceKey,
		Value:   []byte(fmt.Sprintf("%d", nodeConfig.Port)),
		Session: sessionId,
	}
	//start polling to aquire the lock indefinitely
	for {
		if isLeader {
			//if the current node is leader, stop the busy loop for now
			return
		}
		aquired, queryDuration, err := client.KV().Acquire(kvPair, nil)
		if err != nil {
			log.Errorf("Failed to aquire lock", err)
			continue
		}
		if aquired {
			isLeader = aquired
			log.Infof("Lock aquired and marking the node  as leader, Took : %dms\n",
				queryDuration.RequestTime.Milliseconds())
		} else {

			if !loggedOnce {
				log.Debug("Lock is already taken, will check again...")
				loggedOnce = true
			}
			//every 10ms attempt to grab a lock on the consul.
			time.Sleep(10 * time.Millisecond)
		}

	}
}

func InitiateLeaderElection(serverConfig config.Server, nodeId string, sessionId string) {

	go leaderElection(NodeConfig{
		NodeId: nodeId,
		NodeIp: serverConfig.Host,
		Port:   serverConfig.Port,
	}, sessionId, serverConfig.ServiceKey)
}
