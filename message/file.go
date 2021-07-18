package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

//Number is picked from the rust program for copying large files, 8kb seems to perform well.
//Should be benchmarked and updated accordignly if needed.
const DEFAULT_BUFFER_SIZE = 1024 * 8

func getTopicDatDir(dataDir string, topic string) string {
	if !strings.HasSuffix(dataDir, "/") {
		dataDir += "/"
	}
	return dataDir + topic + "/"
}

//check if the directory exist create otherwise
func createPath(path string) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0770)
	}
	return err
}

func writeToLogFile(dataDir string, topic string, messages []Message) (int, error) {

	path := getTopicDatDir(dataDir, topic)
	err := createPath(path)

	if err != nil {
		return 0, err
	}

	fileOffset := messages[0].Id
	path = fmt.Sprintf("%s/%s_%d.log", path, topic, fileOffset)

	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		return 0, err
	}

	//writing 1KB of data took 376microseconds
	//writing 1GB of data took 478milliseconds
	fw := bufio.NewWriter(file)

	write_buffer := make([]byte, 0)
	buf := bytes.NewBuffer(write_buffer)

	counter := 0
	for _, message := range messages {
		counter += 1
		buf.WriteString(fmt.Sprintf("%d=%s", message.Id, message.Data))
	}
	n, err := fw.Write(buf.Bytes())
	if err != nil {
		log.Error().Err(err).Msg("failed to write message buffer to file")
		//todo: should have retrier, cannot afford to loose the messages.
	}

	if n != buf.Len() {
		//FIXME: how to handle this situation
	}
	err = fw.Flush()

	if err != nil {
		log.Error().Err(err).Msg("failed to write message buffer to file")
		//todo: should have retrier, cannot afford to loose the messages.
	}

	return counter, nil
}

func readFromLog(dataDir, topic string, fileOffset int64) ([]Message, error) {
	//path should exist
	path := getTopicDatDir(dataDir, topic)
	path = fmt.Sprintf("%s/%s_%d.log", path, topic, fileOffset)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, DEFAULT_BUFFER_SIZE)
	reader := bufio.NewReader(file)
	var n int
	var bufWriteError error
	//should we create a buffer of fixed size?
	buffer := new(bytes.Buffer)
	for {
		n, err = reader.Read(buf)
		if n == 0 {
			if err == io.EOF {
				err = nil
				break
			} else if err != nil {
				break
			}
		}
		_, bufWriteError = buffer.Write(buf[:n])
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to read log file")
	}
	if bufWriteError != nil {
		return nil, errors.Wrap(bufWriteError, "failed to write log buffer")
	}
	return decodeRawMessage(buffer.Bytes()), nil

}

func decodeRawMessage(raw []byte) []Message {

	messages := make([]Message, 0)
	for _, line := range strings.Split(string(raw), "\n") {
		message := Message{}
		splits := strings.Split(line, "=")
		if len(splits) != 2 {
			continue
		}
		id, err := strconv.ParseUint(splits[0], 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to read message")
			continue
		}
		message.Id = id
		message.Data = splits[1]
		messages = append(messages, message)
	}
	return messages
}
