package process_test

import (
	"fmt"
	"testing"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/filesys"
	"github.com/manderson5192/memfs/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProcessTestSuite struct {
	suite.Suite
	fs filesys.FileSystem
	p  process.ProcessFilesystemContext
}

func (s *ProcessTestSuite) SetupTest() {
	// Setup a process context with a basic file tree
	s.fs = filesys.NewFileSystem()
	s.p = process.NewProcessFilesystemContext(s.fs)
	assert.Nil(s.T(), s.p.MakeDirectory("/a"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/zzz"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b/c"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b/a"))
	foobarFile, err := s.p.CreateFile("/a/foobar_file")
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), foobarFile.TruncateAndWriteAll([]byte("hello!")))
}

func (s *ProcessTestSuite) TestStatRootDir() {
	info, err := s.p.Stat("/")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), directory.FileInfo{
		Size: 1,
		Type: directory.DirectoryType,
	}, *info)
}

func (s *ProcessTestSuite) TestStatOnDir() {
	info, err := s.p.Stat("/a")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), directory.FileInfo{
		Size: 3,
		Type: directory.DirectoryType,
	}, *info)
}

func (s *ProcessTestSuite) TestStatOnFile() {
	info, err := s.p.Stat("/a/foobar_file")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), directory.FileInfo{
		Size: 6,
		Type: directory.FileType,
	}, *info)
}

func (s *ProcessTestSuite) TestStatNoExist() {
	info, err := s.p.Stat("/noexist")
	assert.Nil(s.T(), info)
	assert.NotNil(s.T(), err)
}

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

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessTestSuite))
}
