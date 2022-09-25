package wal

import (
	"bytes"
	"encoding/gob"
)

type Metadata struct {
	Topic   string `json:"topic"`
	WalFile string `json:"wal_file"`
}

// MARKER unique bytes for delimiter data over network
const MARKER_META = 0xAED
const MARKER_FILE = 0xAEF
const MARKER1 = 0xA
const MARKER2 = 0xE
const MARKER_FILE_END = 0xF
const MARKER_META_END = 0xD

func Marker() []byte {
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

func isMarker(b byte) bool {
	switch b {
	case MARKER1:
	case MARKER2:
	case MARKER_FILE_END:
	case MARKER_META_END:
		return true
	default:
		return false
	}

	return false
}
