package cluster

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

//isLeader mark it as true when the current node becomes the leader
var isLeader = false

//connectToLeader Connect this service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func ConnectToLeader(appConfig config.Server, serviceId string, clusteController ClusterController) {

	thisNode := NewNodeConfig(serviceId, appConfig.Host, appConfig.Port)
	requestBody, _ := thisNode.json()
	for {
		//loop if the current node becomes the leader
		log.Infoln("Registering this service as a follower to the cluster leader...")
		//get the leader
		if !isLeader {
			leaderNode, err := clusteController.GetLeader(appConfig.ServiceKey)
			if err != nil {
				log.Errorln(err)
				//TODO: give it a second and loop back??
				time.Sleep(1 * time.Second)
				continue
			}
			//get the leader information and send a follow request.
			leaderEndpoint := fmt.Sprintf("http://%s:%d%s", leaderNode.NodeIp, leaderNode.Port, "/service/api/follower/register")
			resp, err := http.Post(leaderEndpoint, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				log.Errorln("Failed to register with the leader ", err)
				//FIXME : this is possible when the leader elections are still going on and we call GetLeader.
				// should i return or not?
				return
			}
			response, err := ioutil.ReadAll(resp.Body)
			log.Infof("ConnectToLeader Response : %s\n ", string(response))
			break

			//send connect ping to leader
		} else {
			log.Infoln("Im the leader....")
			return
		}
	}
}

func leaderElection(node Node, c ClusterController, serviceKey string) {
	loggedOnce := false
	//start polling to aquire the lock indefinitely
	for {
		if isLeader {
			//if the current node is leader, stop the busy loop for now
			return
		}
		aquired, queryDuration, err := c.AquireLock(node, serviceKey)
		if err != nil {
			log.Errorln("Failed to aquire lock", err)
			continue
		}
		if aquired {
			isLeader = aquired
			log.Infof("Lock aquired and marking the node  as leader, Took : %dms\n",
				queryDuration.Milliseconds())
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

func InitiateLeaderElection(serverConfig config.Server, nodeId string, c ClusterController) {

	go leaderElection(Node{
		NodeId: nodeId,
		NodeIp: serverConfig.Host,
		Port:   serverConfig.Port,
	}, c, serverConfig.ServiceKey)
}
