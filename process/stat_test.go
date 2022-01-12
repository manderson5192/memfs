package process_test

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/stretchr/testify/assert"
)

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

func (s *ProcessTestSuite) TestStatOnDirTrailingSlash() {
	info, err := s.p.Stat("/a/")
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

func (s *ProcessTestSuite) TestStatOnFileTrailingSlash() {
	_, err := s.p.Stat("/a/foobar_file/")
	assert.NotNil(s.T(), err)
}

func (s *ProcessTestSuite) TestStatNoExist() {
	info, err := s.p.Stat("/noexist")
	assert.Nil(s.T(), info)
	assert.NotNil(s.T(), err)
}
