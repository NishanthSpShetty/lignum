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

func makeMessage() Message {
	id := counter.Next()
	return Message{Id: id, Data: fmt.Sprintf("this is message %d", id)}
}

func makeMessages(count int) []Message {
	counter = NewCounter()
	list := make([]Message, count)
	for i := 0; i < count; i++ {
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
		message  *AMessage
		args     args
		getargs  getargs
		expected []Message
	}{
		{name: "Topic gets created for the new topic and message",
			message: &AMessage{
				counter:    NewCounter(),
				messageMap: make(map[string][]Message)},
			args: args{
				topic: "test-new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 1},
			expected: []Message{{Id: 0, Data: "this is test log 001"}},
		},
		{name: "Messages will be appended to existing topic",
			message: &AMessage{
				counter:    NewCounterWithValue(1),
				messageMap: withTopic(makeMessages(1), "test-new"),
			},
			args: args{
				topic: "test-new",
				msg:   "this is test log 001",
			},
			getargs:  getargs{from: 0, to: 2},
			expected: append(makeMessages(1), Message{Id: 1, Data: "this is test log 001"}),
		},
		{name: "new topic will be created along with existing topics",
			message: &AMessage{
				counter:    NewCounterWithValue(1),
				messageMap: withTopic(makeMessages(1), "test-old"),
			},
			args: args{
				topic: "test-new",
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
		message  *AMessage
		expected []Message
	}{
		{
			name:     "returns empty list of messages when range is equal",
			args:     args{from: 1, to: 1},
			message:  &AMessage{},
			expected: []Message{},
		},
		{
			name:     "returns empty list of messages when there are no messages",
			args:     args{from: 1, to: 10},
			message:  &AMessage{},
			expected: []Message{},
		},

		{
			name: "returns list of messages for a given range when there are messages",
			args: args{from: 0, to: 10},
			message: &AMessage{
				messageMap: withTopic(makeMessages(10), "test"),
			},
			expected: makeMessages(10),
		},

		{
			name: "returns list of messages for a given positive range when there are messages",
			args: args{from: 1, to: 10},
			message: &AMessage{
				messageMap: withTopic(makeMessages(10), "test"),
			},
			expected: makeMessages(10)[1:10],
		},

		{
			name: "returns list of messages for a given positive range when there are messages and `to` is less than available messages",
			args: args{from: 1, to: 8},
			message: &AMessage{
				messageMap: withTopic(makeMessages(10), "test"),
			},
			expected: makeMessages(10)[1:8],
		},
		{
			name: "returns list of all messages when range provided is more than available message",
			args: args{from: 0, to: 100},
			message: &AMessage{
				messageMap: withTopic(makeMessages(50), "test"),
			},
			expected: makeMessages(50),
		},
		{
			name: "returns list of messages `from` till end of the message when `to` in range provided is more than available message and from is positive",
			args: args{from: 8, to: 100},
			message: &AMessage{
				messageMap: withTopic(makeMessages(20), "test"),
			},
			expected: makeMessages(20)[8:20],
		},
	}

	for _, tt := range testCases {
		actual := tt.message.Get("test", tt.args.from, tt.args.to)
		assert.Equal(t, tt.expected, actual, "Message.Get: %s", tt.name)
	}
}
