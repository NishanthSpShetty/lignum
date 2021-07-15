package message

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

//Number is picked from the rust program for copying large files, 8kb seems to perform well.
//Should be benchmarked and updated accordignly if needed.
const DEFAULT_BUFFER_SIZE = 1024 * 8

func getTopicDatDir(dataDir string, topic string) string {
	return dataDir + topic + "/"
}

func WriteToLogFile(dataDir string, topic string, messages []Message) (int, error) {

	start_offset := messages[0].Id
	filename := fmt.Sprintf("%s_%d.log", topic, start_offset)
	path := dataDir + "/" + filename
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
