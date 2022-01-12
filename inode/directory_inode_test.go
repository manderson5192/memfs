package inode_test

import (
	"testing"

	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DirectoryInodeSuite struct {
	suite.Suite
	Root *inode.DirectoryInode
	A    *inode.DirectoryInode
	B    *inode.DirectoryInode
	C    *inode.DirectoryInode
}

func (s *DirectoryInodeSuite) SetupTest() {
	// Setup simple directory structure /a/b/c
	s.Root = inode.NewRootDirectoryInode()
	var err error
	s.A, err = s.Root.AddDirectory("a")
	assert.Nil(s.T(), err)
	s.B, err = s.A.AddDirectory("b")
	assert.Nil(s.T(), err)
	s.C, err = s.B.AddDirectory("c")
	assert.Nil(s.T(), err)
}

func (s *DirectoryInodeSuite) TestDirectoryInodeType() {
	assert.Equal(s.T(), inode.InodeDirectory, s.Root.InodeType())
}

func (s *DirectoryInodeSuite) TestSize() {
	assert.Equal(s.T(), 1, s.A.Size())
	assert.Equal(s.T(), 0, s.C.Size())
}

func (s *DirectoryInodeSuite) TestParent() {
	assert.True(s.T(), s.Root == s.Root.Parent())
	assert.True(s.T(), s.A == s.B.Parent())
}

func (s *DirectoryInodeSuite) TestReverseLookup() {
	name, err := s.A.ReverseLookupEntry(s.B)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "b", name)
}

func (s *DirectoryInodeSuite) TestReverseLookupNoExist() {
	name, err := s.C.ReverseLookupEntry(s.B)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "", name)
}

func (s *DirectoryInodeSuite) TestReverseLookupOnSelf() {
	name, err := s.C.ReverseLookupEntry(s.C)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "", name)
}

func (s *DirectoryInodeSuite) TestIsRootDirectory() {
	rootIsRoot := s.Root.IsRootDirectoryInode()
	assert.True(s.T(), rootIsRoot, "the root directory's inode should be identified as a root directory inode")

	subdirIsRoot := s.A.IsRootDirectoryInode()
	assert.False(s.T(), subdirIsRoot, "subdirectory's inode should not be identified as a root directory inode")
}

func (s *DirectoryInodeSuite) TestLookupSubdirectoryEmptyString() {
	lookedUp, err := s.Root.LookupSubdirectory("")
	assert.Nil(s.T(), err)
	assert.True(s.T(), s.Root == lookedUp)
}

func (s *DirectoryInodeSuite) TestLookupSubdirectoryAbsolutePath() {
	lookedUp, err := s.Root.LookupSubdirectory("/")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assert.Nil(s.T(), lookedUp)
}

func (s *DirectoryInodeSuite) TestLookupSubdirectory() {
	lookedUp, err := s.B.LookupSubdirectory("..//../../../..///.//./a/b//c/")
	assert.Nil(s.T(), err)
	assert.True(s.T(), lookedUp == s.C)
}

func TestDirectoryInodeSuite(t *testing.T) {
	suite.Run(t, new(DirectoryInodeSuite))
}
