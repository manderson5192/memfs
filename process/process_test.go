package process_test

import (
	"testing"

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

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessTestSuite))
}
