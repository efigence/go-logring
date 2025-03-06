package logring

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestNextFileName(t *testing.T) {
	dir := t.TempDir()
	lr, err := NewLogring(dir, t.Name(), Config{
		MaxTotalSize: 10 * 1024 * 1024,
		Files:        4,
	})
	require.NoError(t, err)
	filename := lr.nextFileName()
	assert.Equal(t, t.Name()+".01.", filename)
	// timestamps are not nanosecond accurate, need some delay to actually correctly
	// sort by last used
	time.Sleep(time.Millisecond * 10)

	f, err := os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	require.NoError(t, err)
	f.Write([]byte{11})
	f.Sync()
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".02.", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Sync()
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".03.", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".04.", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()

	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".01.", filename)

	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Sync()
	f.Close()

	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".02.", filename)

	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".03.", filename)
}
