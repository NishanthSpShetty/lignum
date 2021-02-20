package message

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/lignum/config"
)

func TestWriteToLogFile(t *testing.T) {
	messageConfig := config.Message{
		MessageFlushIntervalInMilliSeconds: 1,
		MessageDir:                         "temp",
	}
	message := MessageT{
		"foo": "bar",
	}
	err := os.Mkdir(messageConfig.MessageDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Cannot create data directory for the test : %s ", err.Error())
		return
	}
	err = WriteToLogFile(messageConfig, message)
	if err != nil {
		t.Fatalf("Failed to write to log file : %s", err.Error())
	}

	expected := "foo=bar"
	file := "temp/message_001.dat"
	byts, err := ioutil.ReadFile(file)

	if err != nil {
		t.Fatalf("Failed to read log file : %s", err.Error())
		return
	}
	got := string(byts)
	if expected == got {
		t.Fatalf("Got %s, Expected %s", got, expected)
	}
}
