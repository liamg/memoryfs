package memoryfs

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

var separator = "/"

type dir struct {
	info  fileinfo
	dirs  map[string]dir
	files map[string]memfile
}

func (d *dir) Open(name string) (fs.File, error) {

	if name == "" || name == "." {
		return d, nil
	}

	parts := strings.Split(name, separator)

	if len(parts) == 1 {
		if f, ok := d.files[name]; ok {
			f.reader = bytes.NewReader(f.content)
			return &f, nil
		}
	}

	if f, ok := d.dirs[parts[0]]; ok {
		return f.Open(strings.Join(parts[1:], separator))
	}

	return nil, fmt.Errorf("no such file or directory: %s: %w", parts[0], fs.ErrNotExist)
}

func (d *dir) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *dir) find(name string) (fs.FileInfo, error) {

	if name == "" || name == "." {
		return d.info, nil
	}

	parts := strings.Split(name, separator)

	if len(parts) == 1 {
		if f, ok := d.files[name]; ok {
			f.reader = bytes.NewReader(f.content)
			return f.info, nil
		}
	}

	if f, ok := d.dirs[parts[0]]; ok {
		return f.find(strings.Join(parts[1:], separator))
	}

	return nil, fmt.Errorf("no such file or directory: %s: %w", parts[0], fs.ErrNotExist)
}

func (d *dir) ReadDir(name string) ([]fs.DirEntry, error) {

	if name == "" {
		var entries []fs.DirEntry
		for _, file := range d.files {
			entries = append(entries, file.info)
		}
		for _, dir := range d.dirs {
			entries = append(entries, dir.info)
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		return entries, nil
	}

	parts := strings.Split(name, separator)

	dir, ok := d.dirs[parts[0]]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return dir.ReadDir(strings.Join(parts[1:], separator))
}

func (f *dir) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("cannot read directory")
}

func (f *dir) Close() error {
	return nil
}

func (f *dir) MkdirAll(path string, perm fs.FileMode) error {
	parts := strings.Split(path, separator)

	if _, ok := f.files[parts[0]]; ok {
		return fs.ErrExist
	}

	f.info.modified = time.Now()

	if _, ok := f.dirs[parts[0]]; !ok {
		f.dirs[parts[0]] = dir{
			info: fileinfo{
				name:     parts[0],
				size:     0x100,
				modified: time.Now(),
				isDir:    true,
				mode:     perm,
			},
			dirs:  map[string]dir{},
			files: map[string]memfile{},
		}
	}

	if len(parts) == 1 {
		return nil
	}

	dir := f.dirs[parts[0]]
	err := dir.MkdirAll(strings.Join(parts[1:], separator), perm)
	f.dirs[parts[0]] = dir
	return err
}

func (f *dir) WriteFile(path string, data []byte, perm fs.FileMode) error {
	parts := strings.Split(path, separator)

	if len(parts) == 1 {
		buffer := make([]byte, len(data))
		copy(buffer, data)
		f.files[parts[0]] = memfile{
			info: fileinfo{
				name:     parts[0],
				size:     int64(len(buffer)),
				modified: time.Now(),
				isDir:    false,
				mode:     perm,
			},
			content: buffer,
		}
		return nil
	} else if _, ok := f.dirs[parts[0]]; !ok {
		return fs.ErrNotExist
	}
	dir := f.dirs[parts[0]]
	err := dir.WriteFile(strings.Join(parts[1:], separator), data, perm)
	f.dirs[parts[0]] = dir
	return err
}
