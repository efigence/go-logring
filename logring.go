package logring

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Logring interface {
	io.WriteCloser
}

var bufferSize = 65536

type logring struct {
	cfg              *Config
	buf              *bytes.Buffer
	dir              string
	dataSize         int64
	dataCheckSize    int64
	fileSize         int64
	fileCount        int64
	writer           io.WriteCloser
	compressedWriter countingWriter
	lock             sync.Mutex
	prefix           string
	suffix           string
	fd               *os.File
}

type Config struct {
	DirectoryMode fs.FileMode
	FileMode      fs.FileMode
	MaxTotalSize  int64
	Files         int64
}

// Create new log ring within a given directory. Directory will be created with given mode if not existing
func NewLogring(dir string, prefix string, cfg Config) (*logring, error) {
	if cfg.Files <= 1 {
		cfg.Files = 2
	}
	if cfg.MaxTotalSize <= 1024*1024 {
		return nil, fmt.Errorf("max size must be bigger than 1MB")
	}
	if len(dir) == 0 || len(prefix) == 0 {
		return nil, fmt.Errorf("dir and prefix must be longer than 0")
	}
	if cfg.DirectoryMode == 0 {
		cfg.DirectoryMode = 0755
	}
	if cfg.FileMode == 0 {
		cfg.FileMode = 0644
	}
	singleFileSize := cfg.MaxTotalSize / cfg.Files
	lr := logring{
		cfg:       &cfg,
		dir:       dir,
		fileSize:  singleFileSize,
		fileCount: cfg.Files,
		prefix:    prefix,
		buf:       &bytes.Buffer{},
	}
	lr.suffix = "zstd"
	lr.buf.Grow(bufferSize)
	di, _ := os.Stat(lr.dir)
	if !di.IsDir() {
		err := os.MkdirAll(dir, cfg.DirectoryMode)
		if err != nil {
			return nil, fmt.Errorf("could not create/open directory: %s", err)
		}
	}
	_, err := os.ReadDir(lr.dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %s", err)
	}
	fn := dir + "/" + lr.nextFileName()
	lr.fd, err = os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, cfg.FileMode)

	if err != nil {
		return nil, err
	}
	lr.writer, err = zstd.NewWriter(lr.fd)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

func (lr *logring) nextFileName() string {
	files, err := os.ReadDir(lr.dir)
	if err != nil {
		panic("could not read dir")
	}
	re := regexp.MustCompile(regexp.QuoteMeta(lr.prefix) + `\.([0-9]+)\.` + lr.suffix)
	filesInRotation := map[int]string{}
	oldestTS := time.Now()
	var oldestID int
	for _, f := range files {
		matches := re.FindStringSubmatch(f.Name())
		if len(matches) < 2 {
			continue
		}
		fileIDstr := matches[1]
		fileID, _ := strconv.Atoi(fileIDstr)
		filesInRotation[fileID] = f.Name()

		i, _ := f.Info()
		if i.ModTime().Before(oldestTS) {
			oldestTS = i.ModTime()
			oldestID = fileID
		}
	}
	if len(filesInRotation) < int(lr.fileCount) {
		return fmt.Sprintf("%s.%02d.%s", lr.prefix, len(filesInRotation)+1, lr.suffix)
	} else {
		return fmt.Sprintf("%s.%02d.%s", lr.prefix, oldestID, lr.suffix)
	}

}

func (lr *logring) Write(p []byte) (n int, err error) {
	lr.dataSize = lr.dataSize + int64(len(p))
	lr.dataCheckSize = lr.dataCheckSize + int64(len(p))
	if lr.dataCheckSize > lr.fileSize {
		lr.checkForRotate()
	}
	lr.lock.Lock()
	defer lr.lock.Unlock()
	if len(p) <= bufferSize {
		n, err = lr.buf.Write(p)
		if (lr.buf.Len()) > 65535 {
			_, err := lr.buf.WriteTo(lr.writer)
			if err != nil {
				return 0, err
			}
			lr.buf.Reset()
		}
		return n, err
	} else {
		if lr.buf.Len() > 0 {
			_, err := lr.buf.WriteTo(lr.writer)
			if err != nil {
				return 0, err
			}
		}
		return lr.writer.Write(p)
	}
}
func (lr *logring) checkForRotate() {
	fi, err := lr.fd.Stat()
	if err != nil {
		panic(err)
	}
	lr.dataCheckSize = 0
	if fi.Size() > lr.fileSize {
		lr.Rotate()
	}

}

func (lr *logring) Rotate() (err error) {
	lr.lock.Lock()
	defer lr.lock.Unlock()
	old := lr.fd
	defer old.Close()
	fn := lr.dir + "/" + lr.nextFileName()
	lr.fd, err = os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, lr.cfg.FileMode)
	if err != nil {
		return err
	}
	lr.writer, err = zstd.NewWriter(lr.fd)
	lr.dataSize = 0
	lr.dataCheckSize = 0
	return err
}

func (lr *logring) Close() {
	lr.lock.Lock()
	defer lr.lock.Unlock()
	if lr.buf.Len() > 0 {
		lr.buf.WriteTo(lr.writer)
	}
	lr.writer.Close()
}
