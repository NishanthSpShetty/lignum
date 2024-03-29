package types

import (
	"fmt"
	"sync"

	"github.com/NishanthSpShetty/lignum/message/counter"
)

// Message
type Message struct {
	Id   uint64
	Data []byte
}

// Topic
type Topic struct {
	counter       *counter.Counter
	name          string
	messageBuffer []Message
	// number of messages allowed to stay in memory
	msgBufferSize   uint64
	bufferIdx       uint64
	lock            sync.Mutex
	dataDir         string
	liveReplication bool
}

func (m Message) String() string {
	return fmt.Sprintf("{ID: %v, Msg: %s}\n", m.Id, m.Data)
}

type Store interface {
	GetTopics() []*Topic
}
