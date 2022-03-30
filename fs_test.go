package memoryfs

import (
	"io/fs"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Basics(t *testing.T) {

	memfs := New()

	/*map[string][]byte{
		"test.txt":             []byte("hello world"),
		"files/a/b/c/.secret":  []byte("secret file!"),
		"files/a/b/c/note.txt": []byte(":)"),
		"files/a/middle.txt":   []byte(":("),
	})*/

	require.NoError(t, memfs.MkdirAll("files/a/b/c", 0o700))
	require.NoError(t, memfs.WriteFile("test.txt", []byte("hello world"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/b/c/.secret", []byte("secret file!"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/b/c/note.txt", []byte(":)"), 0o644))
	require.NoError(t, memfs.WriteFile("files/a/middle.txt", []byte(":("), 0o644))

	t.Run("Open file", func(t *testing.T) {
		f, err := memfs.Open("test.txt")
		require.NoError(t, err)
		data, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
		require.NoError(t, f.Close())
	})

	t.Run("Open file in dir", func(t *testing.T) {
		f, err := memfs.Open("files/a/b/c/.secret")
		require.NoError(t, err)
		data, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "secret file!", string(data))
		require.NoError(t, f.Close())
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
	})

	t.Run("List directory with file and dir", func(t *testing.T) {
		entries, err := fs.ReadDir(memfs, "files/a")
		require.NoError(t, err)
		require.Len(t, entries, 2)
	})

	t.Run("List directory with multiple files", func(t *testing.T) {
		entries, err := fs.ReadDir(memfs, "files/a/b/c")
		require.NoError(t, err)
		require.Len(t, entries, 2)
	})

	t.Run("Stat root", func(t *testing.T) {
		info, err := memfs.Stat(".")
		require.NoError(t, err)
		assert.Equal(t, ".", info.Name())
		assert.Equal(t, fs.FileMode(0o755), info.Mode())
		assert.Equal(t, true, info.IsDir())
	})

	t.Run("Walk directory", func(t *testing.T) {
		var entries []fs.DirEntry
		err := fs.WalkDir(memfs, ".", func(_ string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			entries = append(entries, info)
			return nil
		})
		require.NoError(t, err)
		require.Len(t, entries, 9)
	})

}
