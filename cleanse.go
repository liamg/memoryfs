package memoryfs

import "strings"

func cleanse(path string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path, "."), separator)
}
