package message

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

/*
How do we structure message file in the message directory?
	1) Create a files with the sequence number each containing specific amount of messages which should be loaded easily
*/

//WriteToLogFile werite the mesages to log file.
// doesnt handle consecutive writes well, meaning it will write the all the messages from the beginning on each write
func WriteToLogFile(messageConfig config.Message, messages []Message) (int, error) {

	filename := "message_001.dat"
	path := messageConfig.MessageDir + "/" + filename
	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		return 0, err
	}

	counter := 0
	for _, message := range messages {
		counter += 1
		file.WriteString(fmt.Sprintf("%d=%s", message.Id, message.Data))
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

//ReadFromLogFile
func ReadFromLogFile(messageDirectory string) []Message {
	//load all mesage files from the given directory,
	filename := "message_001.dat"

	file, err := os.Open(messageDirectory + "/" + filename)

	if err != nil {
		log.Debug().Err(err).Msg("Failed to find message file")
		return nil
	}

	messages, err := ioutil.ReadAll(file)

	if err != nil {
		log.Error().Err(err).Msg("Failed to read the message file")
	}
	return decodeRawMessage(messages)
}
