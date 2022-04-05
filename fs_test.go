package memoryfs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FSInterfaceSupport(t *testing.T) {

	type allFS interface {
		fs.ReadFileFS
		fs.ReadDirFS
		fs.StatFS
		fs.GlobFS
		fs.SubFS
	}

	var m fs.FS = New()
	_, ok := m.(allFS)
	assert.True(t, ok)
}

func Test_AllOperations(t *testing.T) {

	memfs := New()

	require.NoError(t, memfs.MkdirAll("files/a/b/c", 0o700))
	require.NoError(t, memfs.WriteFile("test.txt", []byte("hello world"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/b/c/.secret", []byte("secret file!"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/b/c/note.txt", []byte(":)"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/middle.txt", []byte(":("), 0o644))

	t.Run("Open file", func(t *testing.T) {
		f, err := memfs.Open("test.txt")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		data, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("Open missing file", func(t *testing.T) {
		f, err := memfs.Open("missing.txt")
		require.Error(t, err)
		require.Nil(t, f)
	})

	t.Run("Open directory", func(t *testing.T) {
		f, err := memfs.Open("files")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		require.NotNil(t, f)
		_, err = f.Read([]byte{})
		require.Error(t, err)
	})

	t.Run("Open file in dir", func(t *testing.T) {
		f, err := memfs.Open("files/a/b/c/.secret")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		data, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "secret file!", string(data))
	})

	t.Run("Stat file", func(t *testing.T) {
		info, err := memfs.Stat("test.txt")
		require.NoError(t, err)
		assert.Equal(t, "test.txt", info.Name())
		assert.Equal(t, fs.FileMode(0o644), info.Mode())
		assert.Equal(t, false, info.IsDir())
		assert.Equal(t, int64(11), info.Size())
	})

	t.Run("Stat file in dir", func(t *testing.T) {
		info, err := memfs.Stat("files/a/b/c/.secret")
		require.NoError(t, err)
		assert.Equal(t, ".secret", info.Name())
		assert.Equal(t, fs.FileMode(0o644), info.Mode())
		assert.Equal(t, false, info.IsDir())
		assert.Equal(t, int64(12), info.Size())
	})

	t.Run("Stat missing file", func(t *testing.T) {
		info, err := memfs.Stat("missing.txt")
		require.Error(t, err)
		assert.Nil(t, info)
	})

	t.Run("List directory at root", func(t *testing.T) {
		entries, err := fs.ReadDir(memfs, ".")
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assertEntryFound(t, "files", true, entries)
		assertEntryFound(t, "test.txt", false, entries)
	})

	t.Run("List directory with file and dir", func(t *testing.T) {
		entries, err := fs.ReadDir(memfs, "files/a")
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assertEntryFound(t, "middle.txt", false, entries)
		assertEntryFound(t, "b", true, entries)
	})

	t.Run("List directory with multiple files", func(t *testing.T) {
		entries, err := fs.ReadDir(memfs, "files/a/b/c")
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assertEntryFound(t, ".secret", false, entries)
		assertEntryFound(t, "note.txt", false, entries)
	})

	t.Run("Stat root", func(t *testing.T) {
		info, err := memfs.Stat(".")
		require.NoError(t, err)
		assert.Equal(t, ".", info.Name())
		assert.Equal(t, fs.FileMode(0o700), info.Mode())
		assert.Equal(t, true, info.IsDir())
	})

	t.Run("ReadFile", func(t *testing.T) {
		data, err := memfs.ReadFile("files/a/b/c/note.txt")
		require.NoError(t, err)
		assert.Equal(t, ":)", string(data))
	})

	t.Run("Sub", func(t *testing.T) {
		sub, err := memfs.Sub("files/a")
		require.NoError(t, err)
		data, err := sub.(fs.ReadFileFS).ReadFile("b/c/note.txt")

		require.NoError(t, err)
		assert.Equal(t, ":)", string(data))
	})

	t.Run("Walk directory", func(t *testing.T) {
		assertWalkContainsEntries(
			t,
			memfs,
			".",
			[]string{
				"test.txt",
				"files/a/b/c/.secret",
				"files/a/b/c/note.txt",
				"files/a/middle.txt",
			},
			[]string{
				".",
				"files",
				"files/a",
				"files/a/b",
				"files/a/b/c",
			},
		)
	})

	t.Run("Glob", func(t *testing.T) {
		results, err := memfs.Glob("blah")
		require.NoError(t, err)
		assert.Len(t, results, 0)

		results, err = memfs.Glob("*")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Contains(t, results, "files")
		assert.Contains(t, results, "test.txt")

		results, err = memfs.Glob("files/*/b/*/*.txt")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, results, strings.ReplaceAll("files/a/b/c/note.txt", "/", separator))
	})

	t.Run("Lazy read", func(t *testing.T) {
		err := memfs.WriteLazyFile("files/lazy.txt", func() (io.Reader, error) {
			return strings.NewReader("hello"), nil
		}, 0644)
		require.NoError(t, err)

		data, err := memfs.ReadFile("files/lazy.txt")
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))

		data, err = memfs.ReadFile("files/lazy.txt")
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
	})

	t.Run("Lazy write", func(t *testing.T) {

		buffer := bytes.NewBuffer([]byte{})
		_, err := buffer.Write([]byte{1, 2, 3})
		require.NoError(t, err)

		err = memfs.WriteLazyFile("files/lazy.rw", func() (io.Reader, error) {
			return buffer, nil
		}, 0644)
		require.NoError(t, err)

		data, err := memfs.ReadFile("files/lazy.rw")
		require.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, data)

		err = memfs.WriteFile("files/lazy.rw", []byte{4, 5, 6}, 0644)
		require.NoError(t, err)

		data, err = memfs.ReadFile("files/lazy.rw")
		require.NoError(t, err)
		assert.Equal(t, []byte{4, 5, 6}, data)
	})
}

func Test_ConcurrentWritesToDirectory(t *testing.T) {
	memfs := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			require.NoError(t, memfs.WriteFile(fmt.Sprintf("test_%d.txt", i), []byte("hello world"), 0o644))
			wg.Done()
		}(i)
	}
	wg.Wait()
	entries, err := memfs.ReadDir(".")
	require.NoError(t, err)
	assert.Equal(t, 100, len(entries))
}

func Test_ConcurrentReadsOfFile(t *testing.T) {
	memfs := New()
	err := memfs.WriteFile("test.txt", []byte("hello world"), 0o644)
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := fs.ReadFile(memfs, "test.txt")
			assert.NoError(t, err)
			assert.Equal(t, "hello world", string(data))
		}()
	}
	wg.Wait()
}

func Test_WriteWhileOpen(t *testing.T) {
	memfs := New()
	err := memfs.WriteFile("test.txt", []byte("goodbye world"), 0o644)
	require.NoError(t, err)

	f, err := memfs.Open("test.txt")
	require.NoError(t, err)

	err = memfs.WriteFile("test.txt", []byte("hello world"), 0o644)
	require.NoError(t, err)

	data, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.Equal(t, "hello world", string(data))
}

func Test_MkdirAllRoot(t *testing.T) {
	memfs := New()
	err := memfs.MkdirAll(".", 0o644)
	require.NoError(t, err)
	var count int
	err = fs.WalkDir(memfs, ".", func(_ string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		if count > 1 {
			t.Fatal("more than one file found!")
		}
		return nil
	})
	require.NoError(t, err)
}

type entry struct {
	path string
	info fs.DirEntry
}

func assertEntryFound(t *testing.T, expectedName string, expectedDir bool, entries []fs.DirEntry) {
	var count int
	for _, entry := range entries {
		if entry.Name() == expectedName {
			count++
			if count > 1 {
				t.Errorf("entry %s was found more than once", expectedName)
			}
			assert.Equal(t, expectedDir, entry.IsDir(), "%s was not the expected type", expectedName)
		}
	}
	assert.Greater(t, count, 0, "%s was not found", expectedName)
}

func assertWalkContainsEntries(t *testing.T, f fs.FS, dir string, files []string, dirs []string) {
	var entries []entry
	err := fs.WalkDir(f, dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		entries = append(entries, entry{
			info: info,
			path: path,
		})
		return nil
	})
	require.NoError(t, err)
	for _, expectedFile := range files {
		var count int
		for _, entry := range entries {
			if entry.path == expectedFile {
				count++
				if entry.info.IsDir() {
					t.Errorf("'%s' should be a file, but is a directory", expectedFile)
				}
			}
		}
		assert.Equal(t, 1, count, fmt.Sprintf("file '%s' should have been found once", expectedFile))
	}
	for _, expectedDir := range dirs {
		var count int
		for _, entry := range entries {
			if entry.path == expectedDir {
				count++
				if !entry.info.IsDir() {
					t.Errorf("'%s' should be a file, but is a directory", expectedDir)
				}
			}
		}
		assert.Equal(t, 1, count, fmt.Sprintf("directory '%s' should have been found once", expectedDir))
	}
}
