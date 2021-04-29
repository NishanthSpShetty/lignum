package message

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

/*
How do we structure message file in the message directory?
	1) Create a files with the sequence number each containing specific amount of messages which should be loaded easily
*/

//WriteToLogFile werite the mesages to log file.
// doesnt handle consecutive writes well, meaning it will write the all the messages from the beginning on each write
func WriteToLogFile(messageConfig config.Message, messages []MessageT) (int, error) {

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
		file.WriteString(fmt.Sprintf("%d=%s", message.Id, message.Message))
	}
	return counter, nil
}

func decodeRawMessage(raw []byte) []MessageT {

	messages := make([]MessageT, 0)
	for _, line := range strings.Split(string(raw), "\n") {
		message := MessageT{}
		splits := strings.Split(line, "=")
		if len(splits) != 2 {
			continue
		}
		id, err := strconv.Atoi(splits[0])
		if err != nil {
			log.Errorf("failed to read message :%v\n", err)
			continue
		}
		message.Id = id
		message.Message = splits[1]
		messages = append(messages, message)
	}
	return messages
}

//ReadFromLogFile
func ReadFromLogFile(messageDirectory string) []MessageT {
	//load all mesage files from the given directory,
	filename := "message_001.dat"

	file, err := os.Open(messageDirectory + "/" + filename)

	if err != nil {
		log.Debugf("Failed to find message file : %s", err.Error())
		return nil
	}

	messages, err := ioutil.ReadAll(file)

	if err != nil {
		log.Errorf("Failed to read the message file : %s", err.Error())
	}
	return decodeRawMessage(messages)
}
