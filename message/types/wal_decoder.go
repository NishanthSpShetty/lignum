package types

import (
	"strconv"
	"strings"

	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/rs/zerolog/log"
)

// DecodeRawMessage naively implement the decoding the message written in raw bytes
func DecodeRawMessage(raw []byte, from, to uint64) []*Message {
	buf := make([]*Message, 0, 16)
	i := 0
	for _, line := range strings.Split(string(raw), "\n") {
		message := Message{}
		splits := strings.Split(line, wal.MESSAGE_KEY_VAL_SEPERATOR)
		if len(splits) != 2 {
			continue
		}
		id, err := strconv.ParseUint(splits[0], 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to read message")
			continue
		}

		// skip any messages which arent part of the given range.
		if id < from || id >= to {
			continue
		}

		message.Id = id
		message.Data = splits[1]
		buf = append(buf, &message)
		i += 1
	}
	return buf
}
