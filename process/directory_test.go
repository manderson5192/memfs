package process_test

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestMakeDirectoryWithTrailingSlash() {
	err := s.p.MakeDirectory("/a/b/d/")
	assert.Nil(s.T(), err)

	entries, err := s.p.ListDirectory("/a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "a",
			Type: directory.DirectoryType,
		},
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "d",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *ProcessTestSuite) TestListDirectoryWithTrailingSlash() {
	entries, err := s.p.ListDirectory("/a/b/")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "a",
			Type: directory.DirectoryType,
		},
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *ProcessTestSuite) TestRemoveDirectoryWithTrailingSlash() {
	err := s.p.RemoveDirectory("/a/b/c/")
	assert.Nil(s.T(), err)
	entries, err := s.p.ListDirectory("/a/b/")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "a",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *ProcessTestSuite) TestRemoveDirectoryOnFile() {
	err := s.p.RemoveDirectory("/a/foobar_file")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENotDir)
}

func (s *ProcessTestSuite) TestMakeDirectoryWithAncestorExistingDirectory() {
	err := s.p.MakeDirectoryWithAncestors("/a/b/c")
	assert.Nil(s.T(), err)
}

func (s *ProcessTestSuite) TestMakeDirectoryWithAncestorEntirelyNewDirectory() {
	err := s.p.MakeDirectoryWithAncestors("/x/y/z")
	assert.Nil(s.T(), err)
	info, err := s.p.Stat("/x/y/z")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), directory.DirectoryType, info.Type)
	assert.Equal(s.T(), 0, info.Size)
}

func (s *ProcessTestSuite) TestMakeDirectoryWithAncestorSomeAncestorsExist() {
	err := s.p.MakeDirectoryWithAncestors("/a/b/c/d")
	assert.Nil(s.T(), err)
	info, err := s.p.Stat("/a/b/c/d")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), directory.DirectoryType, info.Type)
	assert.Equal(s.T(), 0, info.Size)
}

func (s *ProcessTestSuite) TestMakeDirectoryWithAncestorAncestorIsFile() {
	err := s.p.MakeDirectoryWithAncestors("/a/foobar_file/subdir")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENotDir)
}
