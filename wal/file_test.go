package wal

import (
	"os"
	"testing"
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
	err := createTestDir("./test/")
	if err != nil {
		t.Fatalf("Cannot create data directory for the test : %s ", err.Error())
	}
}

func TestReadFromLogFile(t *testing.T) {
}
