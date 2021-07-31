package message

import (
	"os"
	"reflect"
	"testing"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message/types"
	"github.com/prometheus/common/log"
)

const TempDirectory = "temp"

func createTestDir(dir string) error {

	err := os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func TestWriteToLogFile(t *testing.T) {
	messageConfig := config.Message{
		InitialSizePerTopic: 10,
		DataDir:             TempDirectory,
	}
	message := types.Message{
		Id:   0,
		Data: "streaming message 1",
	}
	topic := "test_topic"

	err := createTestDir(messageConfig.DataDir)

	if err != nil {
		t.Fatalf("Cannot create data directory for the test : %s ", err.Error())
	}
	count, err := writeToLogFile(messageConfig.DataDir, topic, []types.Message{message})
	if err != nil {
		t.Fatalf("Failed to write to log file : %s", err.Error())
	}
	expectedCount := 1
	if count != expectedCount {
		t.Fatalf("WriteMessageCount: Got %d, Expected %d", count, expectedCount)
	}

	expected := []*types.Message{&message}

	got, err := readFromLog(TempDirectory, "test_topic", 0, 0, 10)

	if err != nil {
		t.Fatal(err)
		return
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got %s, Expected %s", got, expected)
	}
	err = os.RemoveAll(messageConfig.DataDir)
	if err != nil {
		log.Infof("failed to remove test directory, delete it manually. Path : %s", messageConfig.DataDir)
	}
}

func TestReadFromLogFile(t *testing.T) {
}
