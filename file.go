package memoryfs

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"
)

type file struct {
	sync.RWMutex
	info    fileinfo
	content []byte
}

type fileAccess struct {
	file   *file
	buffer io.Writer
	reader io.Reader
}

const bufferSize = 0x100

func (f *file) overwrite(data []byte, perm fs.FileMode) error {
	f.Lock()
	defer f.Unlock()
	if cap(f.content) < len(data) {
		c := len(data)
		if c < bufferSize {
			c = bufferSize
		}
		f.content = make([]byte, len(data), c)
	}
	copy(f.content, data)
	if len(f.content) > len(data) {
		f.content = f.content[:len(data)]
	}
	f.info.size = int64(len(data))
	f.info.modified = time.Now()
	f.info.mode = perm
	return nil
}

func (f *file) openR() *fileAccess {
	f.RLock()
	defer f.RUnlock()
	return &fileAccess{
		file: f,
	}
}

func (f *fileAccess) Stat() (fs.FileInfo, error) {
	f.file.RLock()
	defer f.file.RUnlock()
	return f.file.info, nil
}

func (f *fileAccess) Read(data []byte) (int, error) {
	f.file.RLock()
	defer f.file.RUnlock()
	if f.reader == nil {
		f.reader = bytes.NewReader(f.file.content)
	}
	return f.reader.Read(data)
}

func (f *fileAccess) Close() error {
	return nil
}
