package filepath_test

import (
	"testing"

	"github.com/manderson5192/memfs/filepath"
	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	assert.Equal(t, "", filepath.Join())
	assert.Equal(t, "", filepath.Join(""))
	assert.Equal(t, "/", filepath.Join("/"))
	assert.Equal(t, "foo/bar", filepath.Join("foo", "bar"))
	assert.Equal(t, "foo/bar/", filepath.Join("foo", "bar/"))
	assert.Equal(t, "/foo/bar", filepath.Join("/foo", "bar"))
	assert.Equal(t, "/foo/bar/", filepath.Join("/foo", "bar/"))
	assert.Equal(t, "a/b", filepath.Join("a/", ".", "b"))
	assert.Equal(t, "/a/b", filepath.Join("/../../../../a/b"))
	assert.Equal(t, "a/../b", filepath.Join("a/../b"))
	assert.Equal(t, "/", filepath.Join("/", ".."))
	assert.Equal(t, "/foo/bar/../fizz/buzz/", filepath.Join("///foo/////", "//bar", "../fizz///.///buzz/"))
}

func TestParsePath(t *testing.T) {
	assert.Equal(t, &filepath.PathInfo{
		Entry:      ".",
		ParentPath: ".",
		MustBeDir:  true,
		IsRelative: true,
	}, filepath.ParsePath(""))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      ".",
		ParentPath: "/",
		MustBeDir:  true,
		IsRelative: false,
	}, filepath.ParsePath("/"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "foobar",
		ParentPath: ".",
		MustBeDir:  false,
		IsRelative: true,
	}, filepath.ParsePath("foobar"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "foobar",
		ParentPath: ".",
		MustBeDir:  true,
		IsRelative: true,
	}, filepath.ParsePath("foobar/"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "c",
		ParentPath: "a/b",
		MustBeDir:  false,
		IsRelative: true,
	}, filepath.ParsePath("a////b/./c"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "c",
		ParentPath: "a/b",
		MustBeDir:  true,
		IsRelative: true,
	}, filepath.ParsePath("a/b//c/"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "c",
		ParentPath: "/a/b",
		MustBeDir:  false,
		IsRelative: false,
	}, filepath.ParsePath("/a/b/c"))
	assert.Equal(t, &filepath.PathInfo{
		Entry:      "c",
		ParentPath: "/a/b",
		MustBeDir:  true,
		IsRelative: false,
	}, filepath.ParsePath("/a/b/c/"))
}
