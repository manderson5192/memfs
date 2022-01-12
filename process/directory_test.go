package process_test

import (
	"github.com/manderson5192/memfs/directory"
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
}
