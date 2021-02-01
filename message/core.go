package message

type MessageT map[string]string

var message MessageT

func init() {
	//initialize the message
	message = make(MessageT, 1000)
}

func Put(key, value string) {
	message[key] = value
}

func Get(key string) string {
	v, ok := message[key]

	if ok {
		return v
	} else {
		return ""
	}
}
