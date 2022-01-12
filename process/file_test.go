package process_test

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestCreateFileWithTrailingSlash() {
	_, err := s.p.CreateFile("/filename/")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *ProcessTestSuite) TestOpenFileWithTrailingSlash() {
	_, err := s.p.OpenFile("/a/foobar_file/")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *ProcessTestSuite) TestDeleteFile() {
	err := s.p.DeleteFile("/a/foobar_file")
	assert.Nil(s.T(), err)
	entries, err := s.p.ListDirectory("/a/")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "b",
			Type: directory.DirectoryType,
		},
		{
			Name: "zzz",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *ProcessTestSuite) TestDeleteFileWithTrailingSlash() {
	err := s.p.DeleteFile("/a/foobar_file/")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *ProcessTestSuite) TestDeleteFileOnDirectory() {
	err := s.p.DeleteFile("/a/b")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EIsDir)
}
