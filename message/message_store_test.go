package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var counter *Counter

func withTopic(m []Message, topic string) map[string][]Message {
	return map[string][]Message{topic: m}
}

func createMsgStore(name string, count, msgBufferSize int64) *MessageStore {
	return &MessageStore{
		messageBufferSize: msgBufferSize,
		topic: map[string]*Topic{name: {
			counter:       NewCounterWithValue(uint64(count)),
			messageBuffer: makeMessages(count, msgBufferSize),
			name:          "test_new",
			bufferIdx:     count,
			msgBufferSize: msgBufferSize,
		}},
	}
}

func makeMessage() Message {
	id := counter.Next()
	return Message{Id: id, Data: fmt.Sprintf("this is message %d", id)}
}

func makeMessages(count int64, bufferSize int64) []Message {
	counter = NewCounter()
	list := make([]Message, bufferSize)
	for i := int64(0); i < count; i++ {
		list[i] = makeMessage()
	}
	return list
}

func Test_messagePut(t *testing.T) {

	type args struct {
		topic string
		msg   string
	}

	type getargs struct {
		from uint64
		to   uint64
	}

	testCases := []struct {
		name     string
		message  *MessageStore
		args     args
		getargs  getargs
		expected []Message
	}{
		{name: "Topic gets created for the new topic and message",
			message: &MessageStore{
				messageBufferSize: 10,
				topic:             make(map[string]*Topic)},
			args: args{
				topic: "test_new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 1},
			expected: []Message{{Id: 0, Data: "this is test log 001"}},
		},
		{name: "Messages will be appended to existing topic",
			message: createMsgStore("test_new", 1, 10),
			args: args{
				topic: "test_new",
				msg:   "this is test log 002",
			},
			getargs:  getargs{from: 0, to: 2},
			expected: append(makeMessages(1, 10)[:1], Message{Id: 1, Data: "this is test log 002"}),
		},
		{name: "new topic will be created along with existing topics",

			message: createMsgStore("test-old", 1, 10),
			args: args{
				topic: "test_new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 1},
			expected: []Message{{Id: 0, Data: "this is test log 001"}},
		},
	}

	for _, tt := range testCases {
		tt.message.Put(context.Background(), tt.args.topic, tt.args.msg)
		assert.Equal(t, tt.expected, tt.message.Get(tt.args.topic, tt.getargs.from, tt.getargs.to), "testPut: %s ", tt.name)
	}
}

func Test_messageGet(t *testing.T) {
	type args struct {
		from uint64
		to   uint64
	}

	testCases := []struct {
		name     string
		args     args
		message  *MessageStore
		expected []Message
	}{
		{
			name:     "returns empty list of messages when range is equal",
			args:     args{from: 1, to: 1},
			message:  &MessageStore{},
			expected: []Message{},
		},
		{
			name:     "returns empty list of messages when there are no messages",
			args:     args{from: 1, to: 10},
			message:  &MessageStore{},
			expected: []Message{},
		},

		{
			name:     "returns list of messages for a given range when there are messages",
			args:     args{from: 0, to: 10},
			message:  createMsgStore("test", 10, 10),
			expected: makeMessages(10, 10),
		},

		{
			name:     "returns list of messages for a given positive range when there are messages",
			args:     args{from: 1, to: 10},
			message:  createMsgStore("test", 10, 10),
			expected: makeMessages(10, 10)[1:10],
		},

		{
			name:     "returns list of messages for a given positive range when there are messages and `to` is less than available messages",
			args:     args{from: 1, to: 8},
			message:  createMsgStore("test", 10, 10),
			expected: makeMessages(10, 10)[1:8],
		},
		{
			name:     "returns list of all messages when range provided is more than available message",
			args:     args{from: 0, to: 100},
			message:  createMsgStore("test", 50, 50),
			expected: makeMessages(50, 50),
		},
		{
			name:     "returns list of messages `from` till end of the message when `to` in range provided is more than available message and from is positive",
			args:     args{from: 8, to: 100},
			message:  createMsgStore("test", 20, 20),
			expected: makeMessages(20, 20)[8:20],
		},
	}

	for _, tt := range testCases {
		actual := tt.message.Get("test", tt.args.from, tt.args.to)
		assert.Equal(t, tt.expected, actual, "Message.Get: %s", tt.name)
	}
}
