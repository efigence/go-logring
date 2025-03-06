package logring

import (
	"io"
	"sync/atomic"
)

type countingWriter struct {
	ctr    atomic.Int64
	writer io.Writer
}

func (c *countingWriter) Write(p []byte) (n int, err error) {
	c.ctr.Add(int64(len(p)))
	return c.writer.Write(p)

}
