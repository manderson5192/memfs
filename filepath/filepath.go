package filepath

import "strings"

type PathType int

const (
	PathSeparator string = "/"
)

func IsAbsolutePath(path string) bool {
	return strings.HasPrefix(path, PathSeparator)
}

func IsRelativePath(path string) bool {
	return !IsAbsolutePath(path)
}
