package message

import (
	"context"
	"testing"
)

func Test_messagePut(t *testing.T) {

	msg := "streaming message 1"
	message := &AMessage{
		counter:  NewCounter(),
		messages: make([]Message, 0),
	}
	message.Put(context.Background(), msg)

	id := message.messages[0].Id
	data := message.messages[0].Data
	if data != msg || id != 0 {
		t.Fatalf("expected %+v, got %+v", Message{0, msg}, message.messages[0])
	}
}

func Test_messageGet(t *testing.T) {
}
