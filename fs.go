package memoryfs

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"time"
)

type FS struct {
	dir *dir
}

func New() *FS {
	return &FS{
		dir: &dir{
			info: fileinfo{
				name:     ".",
				size:     0x100,
				modified: time.Now(),
				isDir:    true,
				mode:     0o700,
			},
			dirs:  map[string]*dir{},
			files: map[string]*file{},
		},
	}
}

func (m *FS) Stat(name string) (fs.FileInfo, error) {
	name = cleanse(name)
	if f, err := m.dir.getFile(name); err == nil {
		return f.openR().Stat()
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if f, err := m.dir.getDir(name); err == nil {
		return f.Stat()
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return nil, fmt.Errorf("no such file or directory: %s: %w", name, fs.ErrNotExist)
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

func (m *FS) ReadFile(name string) ([]byte, error) {
	f, err := m.dir.Open(name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func (m *FS) Sub(dir string) (fs.FS, error) {
	d, err := m.dir.getDir(dir)
	if err != nil {
		return nil, err
	}
	return &FS{
		dir: d,
	}, nil
}

func (m *FS) Glob(pattern string) ([]string, error) {
	return m.dir.glob(pattern)
}
