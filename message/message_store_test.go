package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/NishanthSpShetty/lignum/message/counter"
	"github.com/NishanthSpShetty/lignum/message/types"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/stretchr/testify/assert"
)

var mycounter *counter.Counter

func withTopic(m []types.Message, topic string) map[string][]types.Message {
	return map[string][]types.Message{topic: m}
}

func seedMessages(count, bufferSize uint64) []types.Message {
	mycounter = counter.NewCounter()
	list := make([]types.Message, bufferSize)
	for i := uint64(0); i < count; i++ {
		list[i] = *makeMessage()
	}
	return list
}

func createMsgStore(name string, count, msgBufferSize uint64) *MessageStore {
	topic := types.NewTopic(name, msgBufferSize, "")
	topic.SetCounter(count)
	topic.PushAll(makeMessages(count, count))
	return &MessageStore{
		messageBufferSize: msgBufferSize,
		walChannel:        make(chan<- wal.Payload, 100),
		topic:             map[string]*types.Topic{name: topic},
	}
}

func makeMessage() *types.Message {
	id := mycounter.Next()
	return &types.Message{Id: id, Data: []byte(fmt.Sprintf("this is message %d", id))}
}

func makeMessages(count, bufferSize uint64) []*types.Message {
	mycounter = counter.NewCounter()
	list := make([]*types.Message, bufferSize)
	for i := uint64(0); i < count; i++ {
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
		expected []*types.Message
	}{
		{
			name: "Topic gets created for the new topic and message",
			message: &MessageStore{
				messageBufferSize: 10,
				walChannel:        make(chan<- wal.Payload, 10),
				topic:             make(map[string]*types.Topic),
			},
			args: args{
				topic: "test_new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 1},
			expected: []*types.Message{{Id: 0, Data: []byte("this is test log 001")}},
		},
		{
			name:    "Messages will be appended to existing topic",
			message: createMsgStore("test_new", 1, 10),
			args: args{
				topic: "test_new",
				msg:   "this is test log 002",
			},
			getargs:  getargs{from: 0, to: 2},
			expected: append(makeMessages(1, 10)[:1], &types.Message{Id: 1, Data: []byte("this is test log 002")}),
		},
		{
			name:    "new topic will be created along with existing topics",
			message: createMsgStore("test_old", 1, 10),
			args: args{
				topic: "test_new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 1},
			expected: []*types.Message{{Id: 0, Data: []byte("this is test log 001")}},
		},
	}

	for _, tt := range testCases {
		fmt.Println(tt.message.topic)
		tt.message.Put(context.Background(), tt.args.topic, []byte(tt.args.msg))
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
		expected []*types.Message
	}{
		{
			name:     "returns empty list of messages when range is equal",
			args:     args{from: 1, to: 1},
			message:  &MessageStore{},
			expected: nil,
		},
		{
			name:     "returns empty list of messages when there are no messages",
			args:     args{from: 1, to: 10},
			message:  &MessageStore{},
			expected: nil,
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
