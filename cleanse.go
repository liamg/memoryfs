package memoryfs

import (
	"path/filepath"
	"strings"
)

func cleanse(path string) string {
	return strings.TrimPrefix(strings.TrimPrefix(filepath.Clean(path), "."), separator)
}
