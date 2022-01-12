package filepath

import (
	"strings"

	"github.com/manderson5192/memfs/utils"
)

type PathType int

const (
	PathSeparatorRune    rune   = '/'
	PathSeparator        string = string(PathSeparatorRune)
	SelfDirectoryEntry   string = "."
	ParentDirectoryEntry string = ".."
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
// Cut() method (from the path/filepath module).  Candidly, Go's implementation is much more
// efficient -- I just figured it was a stretch to use their implementation for this assignment :).
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

// PathInfo represents a path.  Entry and ParentPath are guaranteed to be non-empty strings such
// that Join(ParentPath, Entry) is equivalent to the original path parsed by ParsePath().
type PathInfo struct {
	Entry      string
	ParentPath string
	// MustBeDir is true if this path definitely refers to a directory.  If it is false then this
	// path could be a file or a directory.
	MustBeDir  bool
	IsRelative bool
}

// ParsePath lexically parses a path into (1) the name of the directory or file at the end of the
// path, (2) the path to the subdirectory containing the aforementioned entry, (3) whether the
// path indicates that the entry name must be a directory (e.g. if it ends with a path separator),
// and (4) whether the path is relative (or absolute).  It stores this information in a PathInfo.
func ParsePath(path string) *PathInfo {
	// Clean the path for convenience
	cleanPath := Clean(path)
	// interpret "" as a reference to the current directory
	if cleanPath == "" {
		return &PathInfo{
			Entry:      SelfDirectoryEntry,
			ParentPath: SelfDirectoryEntry,
			MustBeDir:  true,
			IsRelative: true,
		}
	}
	// special case: "/"
	if cleanPath == "/" {
		return &PathInfo{
			Entry:      SelfDirectoryEntry,
			ParentPath: cleanPath,
			MustBeDir:  true,
			IsRelative: false,
		}
	}
	isRelative := IsRelativePath(cleanPath)
	mustBeDir := strings.HasSuffix(cleanPath, "/")
	if mustBeDir {
		cleanPath = cleanPath[0 : len(cleanPath)-1]
	}
	parentPath, entry, found := utils.RightCut(cleanPath, PathSeparator)
	if !found {
		// There was no path separator, so this is a relative path and the whole path is entry name
		parentPath = SelfDirectoryEntry
	}
	return &PathInfo{
		Entry:      entry,
		ParentPath: parentPath,
		MustBeDir:  mustBeDir,
		IsRelative: isRelative,
	}
}
