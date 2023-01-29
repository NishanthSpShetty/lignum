package counter

import "sync/atomic"

type Counter struct {
	value uint64
}

func NewCounter() *Counter {
	return &Counter{value: 0}
}

func NewCounterWithValue(val uint64) *Counter {
	return &Counter{value: val}
}

func (c *Counter) Next() uint64 {
	for {
		oldValue := atomic.LoadUint64(&c.value)
		if atomic.CompareAndSwapUint64(&c.value, oldValue, oldValue+1) {
			return oldValue
		}
	}
}

func (c *Counter) Set(newValue uint64) {
	for {
		oldValue := atomic.LoadUint64(&c.value)
		if atomic.CompareAndSwapUint64(&c.value, oldValue, newValue) {
			return
		}
	}
}

func (c *Counter) Get() uint64 {
	return c.value
}
