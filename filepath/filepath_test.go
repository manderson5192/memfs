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
