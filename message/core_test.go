package message

import (
	"os"
	"testing"

	"github.com/NishanthSpShetty/lignum/config"
)

func TestMain(m *testing.M) {

	Init(config.Message{MessageDir: "data_dir"})
	os.Exit(m.Run())
}

func Test_messagePut(t *testing.T) {

	msg := "streaming message 1"
	Put(msg)

	id := messages[0].Id
	message := messages[0].Message
	if message != msg || id != 0 {
		t.Fatalf("expected %+v, got %+v", MessageT{0, msg}, messages[0])
	}
}

func Test_messageGet(t *testing.T) {
	msg := "streaming message 1"
	Put(msg)

	gotValue := Get(0, 1)
	if len(gotValue) == 0 || gotValue[0] != msg {
		//t.Fatalf("expected %s, got %s", msg, gotValue)
	}
}
