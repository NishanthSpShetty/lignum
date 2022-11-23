package wal

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/stretchr/testify/assert"
)

func Test_getWalOperation(t *testing.T) {

	wal := New(config.Wal{}, "tmp", make(<-chan Payload))
	payload := Payload{
		Topic: "test_wal",
		Id:    101,
		Data:  "test wal writer",
	}
	writer := wal.getWalWriter(payload)

	assert.NotNil(t, writer, "returns WAL writer ")

	//promote non existing topic
	err := wal.Promote("dead_topic")
	assert.NotNil(t, err, "promoting topic which doesnt have WAL file, returns error")

	//promote existing topic
	err = wal.Promote(payload.Topic)
	assert.Nil(t, err, "successfully promotes WAl file")

	assert.Nil(t, wal.walCache.getFile(payload.Topic), "promoting file should set the WAL file for topic to nil")
	assert.Nil(t, wal.walCache.getWriter(payload.Topic), "promoting file should set the WAL writer for topic to nil")

	//try promoting same topic again without change
	err = wal.Promote(payload.Topic)
	assert.NotNil(t, err, "promoting twice should fail")
	//close when done
	wal.walCache.getFile(payload.Topic).Close()
}

func isFileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func Test_walWrite(t *testing.T) {

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	q := make(chan Payload, 1)
	//file format
	logFileStr := "tmp/%s/%s_%d.%s"

	wal := New(config.Wal{}, "tmp", q)

	payload := Payload{
		Topic: "test_topic",
		Id:    10,
		Data:  "test wal writer",
	}

	ctx := context.Background()
	wal.StartWalWriter(ctx)

	//write payload to queue
	q <- payload
	q <- payload
	payload2 := Payload{Topic: "another_topic", Id: 20, Data: " uneventful event "}
	q <- payload2

	walName := fmt.Sprintf(logFileStr, payload.Topic, payload.Topic, payload.Id, "qwal")
	assert.True(t, isFileExist(walName), "WAL file should be created for the payload topic: %s", payload)
	//second topic wal file
	walName = fmt.Sprintf(logFileStr, payload2.Topic, payload2.Topic, payload2.Id, "qwal")
	assert.True(t, isFileExist(walName), "WAL file should be created for the payload topic %s", payload2)
	//send promote signal
	q <- Payload{Promote: true, Topic: payload.Topic}
	//let the routine to process the message
	time.Sleep(2 * time.Second)
	assert.Nil(t, wal.walCache.getWriter(payload.Topic), "promoting file should set the WAL writer for topic to nil")
	assert.Nil(t, wal.walCache.getFile(payload.Topic), "promoting file should set the WAL file for topic to nil")

	//there will be log file
	fileName := fmt.Sprintf(logFileStr, payload.Topic, payload.Topic, payload.Id, "log")

	payload = Payload{
		Topic: "test_topic",
		Id:    200,
		Data:  "test wal writer after promoting prev file",
	}
	q <- payload
	q <- payload
	//test after writing couple of messages to same topic, promoting file might take some time
	assert.True(t, isFileExist(fileName), "WAL file should be promoted when Promote is signaled in payload")
	q <- Payload{Promote: true, Topic: payload.Topic}
}
