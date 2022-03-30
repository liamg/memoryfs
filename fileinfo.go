package memoryfs

import (
	"io/fs"
	"time"
)

type fileinfo struct {
	name     string
	size     int64
	modified time.Time
	isDir    bool
	mode     fs.FileMode
}

func (f fileinfo) Name() string {
	return f.name
}

func (f fileinfo) Size() int64 {
	return f.size
}

func (f fileinfo) Mode() fs.FileMode {
	return f.mode
}

func (f fileinfo) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f fileinfo) Type() fs.FileMode {
	return f.Mode()
}

func (f fileinfo) ModTime() time.Time {
	return f.modified
}

func (f fileinfo) IsDir() bool {
	return f.isDir
}

func (f fileinfo) Sys() interface{} {
	return nil
}
