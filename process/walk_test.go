package process_test

import (
	"fmt"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/process"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestWalk() {
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		paths = append(paths, path)
		return nil
	})
	err := s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/b",
		"/a/b/a",
		"/a/b/c",
		"/a/foobar_file",
		"/a/zzz",
	}, paths)
}

func (s *ProcessTestSuite) TestWalkWalkFuncReturnsErr() {
	walkFuncErr := fmt.Errorf("this error stops the WalkFunc")
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		if len(paths) >= 3 {
			return walkFuncErr
		}
		paths = append(paths, path)
		return nil
	})
	err := s.p.Walk("/", walkFn)
	assert.Equal(s.T(), walkFuncErr, err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/b",
	}, paths)
}

func (s *ProcessTestSuite) TestWalkWalkFuncSkipsB() {
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		// Skip directory /a/b
		if path == "/a/b" {
			return process.SkipDir
		}
		paths = append(paths, path)
		return nil
	})
	err := s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/foobar_file",
		"/a/zzz",
	}, paths)
}

func (s *ProcessTestSuite) TestWalkWalkFuncSkipsFooBarFile() {
	paths := make([]string, 0)
	walkFn := process.WalkFunc(func(path string, fileInfo *directory.FileInfo, err error) error {
		assert.Nil(s.T(), err, "WalkFunc shouldn't receive any errors")
		assert.NotNil(s.T(), fileInfo, "fileInfo should be populated on all calls to WalkFunc")
		// Skip file /a/foobar_file
		if path == "/a/foobar_file" {
			return process.SkipDir
		}
		paths = append(paths, path)
		return nil
	})
	err := s.p.Walk("/", walkFn)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		"/",
		"/a",
		"/a/b",
		"/a/b/a",
		"/a/b/c",
	}, paths)
}
