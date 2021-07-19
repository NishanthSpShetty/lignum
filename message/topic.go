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
	msgBufferSize uint64
	bufferIdx     uint64
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

func (t *Topic) getBufferStartOffset() uint64 {
	return t.messageBuffer[0].Id
}

func (t *Topic) getFileOffset(id uint64) uint64 {
	if id == 0 || t.msgBufferSize == 0 {
		return 0
	}

	mod := id % t.msgBufferSize
	return id - mod
}

func (t *Topic) readFromBuffer(from, to uint64) []Message {

	//if both offset points to inbuffer messages, read from buffer.
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

//GetMessages get all messages written to the topic.
//If message ranges lie within buffered messages, return them, if not check if it already written to files.
func (t *Topic) GetMessages(from, to uint64) []Message {

	fromOffset := t.getFileOffset(from)
	toOffset := t.getFileOffset(to)

	currentInbufferOffset := t.getBufferStartOffset()
	fromInBuffer := false
	toInBuffer := true

	if fromOffset == currentInbufferOffset {
		//start range is in buffer
		fromInBuffer = true

	}
	if toOffset == currentInbufferOffset {
		//buffer end offset is in buffer,
		toInBuffer = true
	}

	if fromInBuffer && toInBuffer {
		return t.readFromBuffer(from, to)
	}
	return nil
}

func (t *Topic) getMessageSizeInBuffer() uint64 {
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
