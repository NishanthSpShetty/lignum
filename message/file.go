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
const DEFAULT_READ_CHUNK_SIZE = 1024 * 4
const MESSAGE_KEY_VAL_SEPERATOR = "="

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
		buf.WriteString(fmt.Sprintf("%d%s%s", message.Id, MESSAGE_KEY_VAL_SEPERATOR, message.Data))
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
	buf := make([]byte, DEFAULT_READ_CHUNK_SIZE)
	reader := bufio.NewReader(file)
	var n int
	var bufWriteError error

	//chunk size of 8kb does provide improved result, anything less than a page size worsens below num
	//bench: Reading 1GB file
	//using 1GB buffer size
	//took ~255ms to 295ms
	//using 1Mb buffer
	//took 370ms to 690ms

	//bench: Reading 100MB file
	//using 1MB buffer size
	//took ~50ms to ~70ms
	//using 1Kb buffer
	//took ~50ms to ~80ms

	//Refer : https://github.com/NishanthSpShetty/bench_disk_io.go

	//TODO: look into reuse of this buffer
	buffer := bytes.NewBuffer(make([]byte, 1024*1024))
	//calling reset will make buffer.Write reuse the underlying buffer.
	//if not called, on write it will grow the underlying buffer causing allocation
	buffer.Reset()
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
		splits := strings.Split(line, MESSAGE_KEY_VAL_SEPERATOR)
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
