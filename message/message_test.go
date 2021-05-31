package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const DummyTopic = "dummy-topic"

func Test_messagePut(t *testing.T) {

	msg := "streaming message 1"
	message := &AMessage{
		counter:    NewCounter(),
		messageMap: make(map[string][]Message, 1),
	}
	message.Put(context.Background(), DummyTopic, msg)

	id := message.messageMap[DummyTopic][0].Id
	data := message.messageMap[DummyTopic][0].Data
	if data != msg || id != 0 {
		t.Fatalf("expected %+v, got %+v", Message{0, msg}, message.GetMessages(DummyTopic)[0])
	}
}

var counter *Counter

func withDummyTopic(m []Message) map[string][]Message {
	return map[string][]Message{DummyTopic: m}
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
				messageMap: withDummyTopic(makeMessages(10)),
			},
			expected: makeMessages(10),
		},

		{
			name: "returns list of messages for a given positive range when there are messages",
			args: args{from: 1, to: 10},
			message: &AMessage{
				messageMap: withDummyTopic(makeMessages(10)),
			},
			expected: makeMessages(10)[1:10],
		},

		{
			name: "returns list of messages for a given positive range when there are messages and `to` is less than available messages",
			args: args{from: 1, to: 8},
			message: &AMessage{
				messageMap: withDummyTopic(makeMessages(10)),
			},
			expected: makeMessages(10)[1:8],
		},
		{
			name: "returns list of all messages when range provided is more than available message",
			args: args{from: 0, to: 100},
			message: &AMessage{
				messageMap: withDummyTopic(makeMessages(50)),
			},
			expected: makeMessages(50),
		},
		{
			name: "returns list of messages `from` till end of the message when `to` in range provided is more than available message and from is positive",
			args: args{from: 8, to: 100},
			message: &AMessage{
				messageMap: withDummyTopic(makeMessages(20)),
			},
			expected: makeMessages(20)[8:20],
		},
	}

	for _, tt := range testCases {

		actual := tt.message.Get(DummyTopic, tt.args.from, tt.args.to)

		assert.Equal(t, tt.expected, actual, "Message.Get: %s", tt.name)
	}
}
