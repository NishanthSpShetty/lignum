package message

import (
	"fmt"
	"sync"
)

type Message struct {
	Id uint64
	//TODO: consider []byte here
	Data string
}

func (m Message) String() string {
	return fmt.Sprintf("{ID: %v, Msg: %s}\n", m.Id, m.Data)
}

type Topic struct {
	counter       *Counter
	name          string
	messageBuffer []Message
	//number of messages allowed to stay in memory
	msgBufferSize int64
	bufferIdx     int64
	lock          sync.Mutex
}

func (t *Topic) GetName() string {
	return t.name
}

func (t *Topic) GetCurrentOffset() uint64 {
	return t.counter.value
}

func (t *Topic) getBufferedMessage() []Message {
	return t.messageBuffer[:t.bufferIdx]
}

func (t *Topic) GetMessages(from, to uint64) []Message {

	msgLen := uint64(t.getMessageSizeInBuffer())
	if msgLen == 0 {
		return nil
	}
	if to > msgLen {
		to = msgLen
	}
	msgs := make([]Message, to-from)
	i := 0
	messages := t.getBufferedMessage()
	for _, msg := range messages {
		if msg.Id < from {
			continue
		}
		if msg.Id >= to {
			break
		}
		msgs[i] = msg
		i++
	}
	return msgs
}

func (t *Topic) getMessageSizeInBuffer() int64 {
	return t.bufferIdx
}

func (t *Topic) resetMessageBuffer() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.messageBuffer = make([]Message, t.msgBufferSize)
	t.bufferIdx = 0
}

func (t *Topic) Push(msg string) Message {
	message := Message{Id: t.counter.Next(), Data: msg}

	t.lock.Lock()
	t.messageBuffer[t.bufferIdx] = message
	t.bufferIdx++
	t.lock.Unlock()
	return message

}
