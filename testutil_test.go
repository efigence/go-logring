package logring

import (
	"bytes"
	"fmt"
	"os"
)

var logDataStore []byte

func logData() []byte {
	if len(logDataStore) == 0 {
		var err error
		logDataStore, err = os.ReadFile("t-data/test.log")
		if err != nil {
			panic(fmt.Sprintf("error loading test log data from t-data/test.log: %s", err))
		}
	}
	return bytes.Clone(logDataStore)
}

func NewCountingDiscardWriter() *discard {
	return &discard{}
}

type discard struct {
	ctr int
}

func (d *discard) Write(p []byte) (int, error) {
	d.ctr += len(p)
	return len(p), nil
}

func (d *discard) WriteString(s string) (int, error) {
	d.ctr += len(s)
	return len(s), nil
}
