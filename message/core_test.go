package message

import (
	"os"
	"testing"

	"github.com/lignum/config"
)

func TestMain(m *testing.M) {

	Init(config.Message{MessageDir: "data_dir"})
	os.Exit(m.Run())
}

func Test_messagePut(t *testing.T) {

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
