# memoryfs

An in-memory filesystem implementation of io/fs.FS.

`memoryfs` implements the following `fs` interfaces:

- [fs.FS](https://pkg.go.dev/io/fs#FS)
- [fs.ReadDirFS](https://pkg.go.dev/io/fs#ReadDirFS)
- [fs.StatFS](https://pkg.go.dev/io/fs#StatFS)

It also allows the creation of files and directories.

## Example

```go
package main

import (
    "fmt"
    "io/fs"

    "github.com/liamg/memoryfs"
)

func main() {

    memfs := memoryfs.New()

    if err := memfs.MkdirAll("my/dir", 0o700); err != nil {
        panic(err)
    }

    if err := memfs.WriteFile("my/dir/file.txt", []byte("hello world"), 0o600); err != nil {
        panic(err)
    }

    data, err := fs.ReadFile(memfs, "my/dir/file.txt")
    if err != nil {
        panic(err)
    }

    fmt.Println(string(data))
}
```
