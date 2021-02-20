package message

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
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
		MessageFlushIntervalInMilliSeconds: 1,
		MessageDir:                         TempDirectory,
	}
	message := MessageT{
		"foo": "bar",
	}

	err := createTestDir(messageConfig.MessageDir)

	if err != nil {
		t.Fatalf("Cannot create data directory for the test : %s ", err.Error())
	}
	count, err := WriteToLogFile(messageConfig, message)
	if err != nil {
		t.Fatalf("Failed to write to log file : %s", err.Error())
	}
	expectedCount := 1
	if count != expectedCount {
		t.Fatalf("WriteMessageCount: Got %d, Expected %d", count, expectedCount)
	}

	expected := "foo=bar"
	file := TempDirectory + "/message_001.dat"
	byts, err := ioutil.ReadFile(file)

	if err != nil {
		t.Fatalf("Failed to read log file : %s", err.Error())
		return
	}
	got := string(byts)
	if expected == got {
		t.Fatalf("Got %s, Expected %s", got, expected)
	}
	err = os.RemoveAll(messageConfig.MessageDir)
	if err != nil {
		log.Infof("failed to remove test directory, delete it manually. Path : %s", messageConfig.MessageDir)
	}
}

func TestReadFromLogFile(t *testing.T) {

	err := createTestDir(TempDirectory)

	if err != nil {
		t.Fatalf("Cannot create data directory for the test : %s ", err.Error())
	}

	file := TempDirectory + "/message_001.dat"
	messageToWrite := "foo=bar"
	ioutil.WriteFile(file, []byte(messageToWrite), os.ModePerm)
	got := ReadFromLogFile(TempDirectory)

	expected := MessageT{
		"foo": "bar",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got %v, Expected %v", got, expected)
	}

	err = os.RemoveAll(TempDirectory)
	if err != nil {
		log.Infof("failed to remove test directory, delete it manually. Path : %s", TempDirectory)
	}
}
