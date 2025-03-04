package logring

import "io"

type Logring interface {
	io.WriteCloser
}

type logring struct {
	buffer []byte
}
