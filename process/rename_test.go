package process_test

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/process"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestRenameAbsoluteAndAbsolutePaths() {
	err := s.p.Rename("/a/b", "/a/new_b")
	assert.Nil(s.T(), err)

	// Verify the resultant tree by Walk()'ing
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		paths = append(paths, path)
		return nil
	})
	err = s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/foobar_file",
		"/a/new_b",
		"/a/new_b/a",
		"/a/new_b/c",
		"/a/zzz",
	}, paths)
}

func (s *ProcessTestSuite) TestRenameRelativeAndRelativePaths() {
	err := s.p.ChangeDirectory("a")
	assert.Nil(s.T(), err)
	err = s.p.Rename("b", "../a/new_b")
	assert.Nil(s.T(), err)

	// Verify the resultant tree by Walk()'ing
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		paths = append(paths, path)
		return nil
	})
	err = s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/foobar_file",
		"/a/new_b",
		"/a/new_b/a",
		"/a/new_b/c",
		"/a/zzz",
	}, paths)
}

func (s *ProcessTestSuite) TestRenameMixedPaths() {
	err := s.p.ChangeDirectory("a")
	assert.Nil(s.T(), err)
	err = s.p.Rename("/a/b", "../a/new_b")
	assert.Nil(s.T(), err)

	// Verify the resultant tree by Walk()'ing
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		paths = append(paths, path)
		return nil
	})
	err = s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/foobar_file",
		"/a/new_b",
		"/a/new_b/a",
		"/a/new_b/c",
		"/a/zzz",
	}, paths)
}
