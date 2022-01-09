package utils

import "strings"

// Cut slices s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, "", false.
//
// NOTE: This method is *almost* available in a release version of Go, but as of this date is still
// only available in go1.18beta1.  For convenience, I have copied the source from:
// https://cs.opensource.google/go/go/+/refs/tags/go1.18beta1:src/strings/strings.go;l=1177
func Cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func RightCut(s, sep string) (before, after string, found bool) {
	if i := strings.LastIndex(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return "", s, false
}
