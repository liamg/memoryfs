package memoryfs

import (
	"io"
	"io/fs"
)

type file struct {
	info    fileinfo
	reader  io.Reader
	content []byte
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *file) Read(data []byte) (int, error) {
	return f.reader.Read(data)
}

func (f *file) Close() error {
	return nil
}
