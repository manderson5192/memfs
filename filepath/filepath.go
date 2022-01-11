package filepath

import "strings"

type PathType int

const (
	PathSeparatorRune rune   = '/'
	PathSeparator     string = string(PathSeparatorRune)
)

func IsAbsolutePath(path string) bool {
	return strings.HasPrefix(path, PathSeparator)
}

func IsRelativePath(path string) bool {
	return !IsAbsolutePath(path)
}

// Clean lexically simplifies a path by applying the following operations, in order:
//	(1) replaces sequential path separators with a single path separator
//	(2) removes '.' entries in the path
//	(3) removes leading sequences of '..' parts from paths that start from '/'
//
// The contract (but not the implementation) of Clean() is inspired by the Go standard library's
// Cut() method (from the path/filepath module)
func Clean(path string) string {
	// Replace sequential path separators with a single path separator
	var builder strings.Builder
	lastRuneWasSeparator := false
	for _, r := range path {
		if r == PathSeparatorRune && lastRuneWasSeparator {
			continue
		}
		if r != PathSeparatorRune && lastRuneWasSeparator {
			lastRuneWasSeparator = false
		}
		lastRuneWasSeparator = PathSeparatorRune == r
		builder.WriteRune(r)
	}
	path = builder.String()

	// Remove '.' elements from the path
	parts := strings.Split(path, PathSeparator)
	sanitizedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "." {
			sanitizedParts = append(sanitizedParts, part)
		}
	}
	path = strings.Join(sanitizedParts, PathSeparator)

	// Remove leading '..' entries from absolute paths
	for IsAbsolutePath(path) && strings.HasPrefix(path, "/../") {
		path = "/" + strings.TrimPrefix(path, "/../")
	}
	if path == "/.." {
		path = "/"
	}

	return path
}

// Join joins together all of the supplied path parts with the PathSeparator before Clean()'ing and
// returning the result
func Join(parts ...string) string {
	return Clean(strings.Join(parts, PathSeparator))
}
