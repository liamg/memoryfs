package memoryfs

import (
	"strings"
)

func cleanse(path string) string {
	path = strings.TrimPrefix(path, separator)
	var accepted []string
	for _, part := range strings.Split(path, separator) {
		switch {
		case part == "" || part == ".":
			continue
		case part == "..":
			if len(accepted) > 0 {
				accepted = accepted[:len(accepted)-1]
			}
			continue
		}
		accepted = append(accepted, part)
	}
	return strings.Join(accepted, separator)
}
