package wal

import (
	"bytes"
	"encoding/gob"
)

type Metadata struct {
	Topic      string `json:"topic"`
	WalFile    string `json:"wal_file"`
	NextOffSet uint64 `json:"next_offset"`
}

// MARKER unique bytes for delimiter data over network
const (
	MARKER_META     = 0xAED
	MARKER_FILE     = 0xAEF
	MARKER1         = 0xA
	MARKER2         = 0xE
	MARKER_FILE_END = 0xF
	MARKER_META_END = 0xD
)

func MetaMarker() []byte {
	return []byte{MARKER1, MARKER2, MARKER_META_END}
}

func FileMarker() []byte {
	return []byte{MARKER1, MARKER2, MARKER_FILE_END}
}

func (m Metadata) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func ToMeta(b []byte) (Metadata, error) {
	reader := bytes.NewReader(b)
	enc := gob.NewDecoder(reader)
	m := &Metadata{}
	err := enc.Decode(m)
	if err != nil {
		return Metadata{}, err
	}

	return *m, nil
}

func IsMarker(b byte) bool {
	switch b {
	case MARKER1, MARKER2, MARKER_FILE_END, MARKER_META_END:
		return true
	default:
		return false
	}
}
