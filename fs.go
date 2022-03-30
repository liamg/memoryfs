package memoryfs

import (
	"io/fs"
	"time"
)

type FS struct {
	dir dir
}

func New() *FS {
	return &FS{
		dir: dir{
			info: fileinfo{
				name:     ".",
				size:     0x100,
				modified: time.Now(),
				isDir:    true,
				mode:     0o755,
			},
			dirs:  map[string]dir{},
			files: map[string]memfile{},
		},
	}
}

func (m *FS) Stat(name string) (fs.FileInfo, error) {
	return m.dir.find(cleanse(name))
}

func (m *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return m.dir.ReadDir(cleanse(name))
}

func (m *FS) Open(name string) (fs.File, error) {
	return m.dir.Open(cleanse(name))
}

func (m *FS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return m.dir.WriteFile(cleanse(path), data, perm)
}

func (m *FS) MkdirAll(path string, perm fs.FileMode) error {
	return m.dir.MkdirAll(cleanse(path), perm)
}
