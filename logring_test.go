package logring

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
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
	assert.Equal(t, t.Name()+".02.zstd", filename)
	// timestamps are not nanosecond accurate, need some delay to actually correctly
	// sort by last used
	time.Sleep(time.Millisecond * 10)

	f, err := os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	require.NoError(t, err)
	f.Write([]byte{11})
	f.Sync()
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".03.zstd", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Sync()
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".04.zstd", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".01.zstd", filename)
	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()

	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".02.zstd", filename)

	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Sync()
	f.Close()

	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".03.zstd", filename)

	time.Sleep(time.Millisecond * 10)

	f, _ = os.OpenFile(dir+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	f.Write([]byte{11})
	f.Close()
	filename = lr.nextFileName()
	assert.Equal(t, t.Name()+".04.zstd", filename)
}

func TestLogring_Write(t *testing.T) {
	dir := t.TempDir()
	lr, err := NewLogring(dir, t.Name(), Config{
		MaxTotalSize: 20 * 1024 * 1024,
		Files:        4,
	})
	require.NoError(t, err)
	for i := 0; i < 10000000; i++ {
		lr.Write([]byte(strconv.Itoa(i) + " - 1\n"))
	}
	lr.Rotate()
	for i := 0; i < 100000000; i++ {
		lr.Write([]byte(strconv.Itoa(i) + " - 2\n"))
	}
	lr.Close()
	dirs, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, dirs, 4)

}
