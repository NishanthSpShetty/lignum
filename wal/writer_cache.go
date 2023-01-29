package wal

import (
	"bufio"
	"os"
	"sync"
)

const (
	UNLOCKED = uint64(0)
	LOCKED   = uint64(1)
)

type walFile struct {
	file   *os.File
	writer *bufio.Writer
}

type walWriterCache struct {
	walFile map[string]*walFile
	state   sync.Mutex
}

func newCache() *walWriterCache {
	return &walWriterCache{
		walFile: make(map[string]*walFile),
		state:   sync.Mutex{},
	}
}

func (w *walWriterCache) get(topic string) (*walFile, bool) {
	w.state.Lock()
	defer w.state.Unlock()
	wf, ok := w.walFile[topic]
	return wf, ok
}

func (w *walWriterCache) set(topic string, file *walFile) {
	w.state.Lock()
	defer w.state.Unlock()
	w.walFile[topic] = file
}

func (w *walWriterCache) getWriter(topic string) *bufio.Writer {
	wf, ok := w.get(topic)
	if ok {
		return wf.writer
	}
	return nil
}

func (w *walWriterCache) getFile(topic string) *os.File {
	wf, ok := w.get(topic)
	if ok {
		return wf.file
	}
	return nil
}

func (w *walWriterCache) delete(topic string) {
	w.state.Lock()
	defer w.state.Unlock()
	delete(w.walFile, topic)
}
