package message

import (
	"testing"

	"github.com/lignum/config"
)

func Test_messagePut(t *testing.T) {
	//this should be in some sort of setup call and rest of the test functions should use
	Init(config.Message{InitialSize: 10, MessageDir: "data_dir"})

	key := "messageKey"
	value := "messageValue"
	Put(key, value)

	if gotValue := message[key]; gotValue != value {
		t.Fatalf("expected %s, got %s", value, gotValue)
	}
}

func Test_messageGet(t *testing.T) {

	key := "messageKey"
	value := "messageValue"
	Put(key, value)

	if gotValue := Get(key); gotValue != value {
		t.Fatalf("expected %s, got %s", value, gotValue)
	}
}
