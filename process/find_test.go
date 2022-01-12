package process_test

import (
	"github.com/manderson5192/memfs/fserrors"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestFindAll() {
	paths, err := s.p.FindAll(".", "a")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"a", "a/b/a"}, paths)
}

func (s *ProcessTestSuite) TestFindFirstMatchingFile() {
	path, err := s.p.FindFirstMatchingFile("/", "foo.*")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/a/foobar_file", path)
}

func (s *ProcessTestSuite) TestFindFirstMatchingFileNoMatchExists() {
	path, err := s.p.FindFirstMatchingFile("/", "^a.*")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENoEnt)
	assert.Equal(s.T(), "", path)
}

func (s *ProcessTestSuite) TestFindFirstMatchingFileInvalidPath() {
	path, err := s.p.FindFirstMatchingFile("/path/does/not/exist", "foobar.*")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "", path)
}
