package message

import (
	"io/ioutil"
	"os"
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
func WriteToLogFile(messageConfig config.Message, message MessageT) error {

	filename := "message_001.dat"
	path := messageConfig.MessageDir + "/" + filename
	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		return err
	}

	counter := 0
	for key, value := range message {
		counter += 1
		file.WriteString(key + "=" + value + "\n")
	}
	log.Infof("Written %d messages to file ", counter)
	return nil
}

func decodeRawMessage(messages []byte) MessageT {

	message := make(MessageT)
	for _, line := range strings.Split(string(messages), "\n") {
		splits := strings.Split(line, "=")
		if len(splits) != 2 {
			continue
		}
		key := splits[0]
		value := splits[1]
		message[key] = value
	}
	return message
}

//ReadFromLogFile
func ReadFromLogFile(messageDirectory string) MessageT {
	//load all mesage files from the given directory,
	filename := "message_001.dat"

	file, err := os.Open(messageDirectory + "/" + filename)

	if err != nil {
		log.Debugf("Failed to find message file : %s", err.Error())
		return MessageT{}
	}

	messages, err := ioutil.ReadAll(file)

	if err != nil {
		log.Errorf("Failed to read the message file : %s", err.Error())
	}
	return decodeRawMessage(messages)
}
