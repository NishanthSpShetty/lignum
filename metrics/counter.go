package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var topicCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lignum_total_topics_created",
	Help: "Total number of message topics created in the lignum",
})

var httpPostRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lignum_post_message_request",
	Help: "number of post request",
})

var httpReplicateRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "lignum_replicate_message_request",
	Help: "number of replication message request",
})

var TopicMessageCounter = make(map[string]prometheus.Counter)

func IncrementTopic() {
	topicCounter.Inc()
}

func IncrementPostRequest() {
	httpPostRequestCounter.Inc()
}

func IncrementReplicationRequest() {
	httpReplicateRequestCounter.Inc()
}

func IncrementMessageCount(topic string) {
	counter, ok := TopicMessageCounter[topic]

	if !ok {
		//create new counter
		counter = promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("lignum_%s_message_count", topic),
			Help: "Total number of messages in topic",
		})
		TopicMessageCounter[topic] = counter
	}
	counter.Inc()
}
