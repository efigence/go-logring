package logring

import (
	"bytes"
	"fmt"
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

type logring struct {
	buf              bytes.Buffer
	dir              string
	fileSize         int64
	fileCount        int64
	writer           io.Writer
	compressedWriter countingWriter
	lock             sync.Mutex
	prefix           string
	suffix           string
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
		dir:       dir,
		fileSize:  singleFileSize,
		fileCount: cfg.Files,
		prefix:    prefix,
	}
	di, _ := os.Stat(dir)
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
	lr.lock.Lock()
	defer lr.lock.Unlock()
	n, err = lr.writer.Write(lr.buf.Bytes())
	if (lr.buf.Len() + len(p)) > 65535 {
		lr.buf.WriteTo(lr.writer)
		lr.buf.Reset()
	}
	return n, err

}
