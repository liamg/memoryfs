package memoryfs

import (
	"io"
	"io/fs"
)

type memfile struct {
	info    fileinfo
	reader  io.Reader
	content []byte
}

func (f *memfile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *memfile) Read(data []byte) (int, error) {
	return f.reader.Read(data)
}

func (f *memfile) Close() error {
	return nil
}
