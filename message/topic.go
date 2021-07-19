package message

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
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
	dataDir       string
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
	msgLen := t.getMessageSizeInBuffer()
	if msgLen == 0 {
		return nil
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

func (t *Topic) readFromLogs(fromOffset, toOffset, from, to uint64) []Message {

	msgs := make([]Message, to-from)
	for eachOffset := fromOffset; eachOffset <= toOffset; eachOffset += t.msgBufferSize {
		msg, err := readFromLog(t.dataDir, t.name, eachOffset)
		if err != nil {
			log.Error().Err(err).Msg("failed to read message from the log files")
			return nil
		}
		msgs = append(msgs, msg...)
	}
	return msgs
}

//GetMessages get all messages written to the topic.
//If message ranges lie within buffered messages, return them, if not check if it already written to files.
func (t *Topic) GetMessages(from, to uint64) []Message {

	latestMessageOffset := t.GetCurrentOffset()
	if to > latestMessageOffset {
		to = latestMessageOffset
	}

	fromOffset := t.getFileOffset(from)
	toOffset := t.getFileOffset(to)

	currentInbufferOffset := t.getBufferStartOffset()
	fromInBuffer := false
	toInBuffer := false

	if fromOffset == currentInbufferOffset {
		//start range is in buffer
		fromInBuffer = true

	}
	if toOffset == currentInbufferOffset {
		//buffer end offset is in buffer,
		toInBuffer = true
	}

	log.Debug().
		Uint64("from", from).
		Uint64("to", to).
		Uint64("fromOffset", fromOffset).
		Uint64("toOffset", toOffset).
		Bool("fromInBuffer", fromInBuffer).
		Bool("toInBuffer", toInBuffer).
		Uint64("currentInbufferOffset", currentInbufferOffset).
		Uint64("latestMessageOffset", latestMessageOffset).
		Msg("get stat")

	//if fromInBuffer is true, then all messages can only be read from the buffer.
	if fromInBuffer {
		return t.readFromBuffer(from, to)
	}
	//if toOffset points to in buffer messages, adjust the toFileOffset
	if toInBuffer {
		toOffset = toOffset - t.msgBufferSize
	}
	//read range can start from the file and end at reading buffered message
	var msgs []Message
	if !fromInBuffer {
		msgs = t.readFromLogs(fromOffset, toOffset, from, to)
	}

	if toInBuffer {
		msg := t.readFromBuffer(currentInbufferOffset, to)
		msgs = append(msgs, msg...)
	}
	return msgs
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
