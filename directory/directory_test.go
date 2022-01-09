package directory_test

import (
	"testing"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DirectoryTestSuite struct {
	suite.Suite
	RootDirInode *inode.DirectoryInode
	RootDir      directory.Directory
	ASubdirInode *inode.DirectoryInode
	ASubdir      directory.Directory
	BSubdirInode *inode.DirectoryInode
	BSubdir      directory.Directory
	CSubdirInode *inode.DirectoryInode
	CSubdir      directory.Directory
}

func (s *DirectoryTestSuite) SetupTest() {
	// Create a basic directory tree representing /a/b/c
	s.RootDirInode = inode.NewRootDirectoryInode()
	s.ASubdirInode = addSubdirectory(s.T(), s.RootDirInode, "a")
	s.BSubdirInode = addSubdirectory(s.T(), s.ASubdirInode, "b")
	s.CSubdirInode = addSubdirectory(s.T(), s.BSubdirInode, "c")
	s.RootDir = directory.NewDirectory(s.RootDirInode)
	s.ASubdir = directory.NewDirectory(s.ASubdirInode)
	s.BSubdir = directory.NewDirectory(s.BSubdirInode)
	s.CSubdir = directory.NewDirectory(s.CSubdirInode)

	// Add some additional directories
	addSubdirectory(s.T(), s.RootDirInode, "fizz")
	addSubdirectory(s.T(), s.RootDirInode, "buzz")
	addSubdirectory(s.T(), s.BSubdirInode, "foobar")
}

func addSubdirectory(t *testing.T, d *inode.DirectoryInode, name string) *inode.DirectoryInode {
	newInode, err := d.AddDirectory(name)
	assert.Nil(t, err)
	return newInode
}

func (s *DirectoryTestSuite) TestEquals() {
	aDirOther := directory.NewDirectory(s.ASubdirInode)

	// // Then equality should be evaluated based on underlying inodes
	assert.True(s.T(), s.RootDir.Equals(s.RootDir), "root dir is .Equal() to itself")
	assert.True(s.T(), s.ASubdir.Equals(s.ASubdir), "'a' subdir is .Equal() to itself")
	assert.True(s.T(), s.ASubdir.Equals(aDirOther), "'a' subdir is .Equal() to an equivalent Directory but separate")
	assert.True(s.T(), aDirOther.Equals(s.ASubdir), "'a' subdir is .Equal() to an equivalent Directory but separate (opposite order)")
	assert.False(s.T(), s.ASubdir == aDirOther, "'a' subdir is not == to a .Equal() but separate Directory")
	assert.False(s.T(), s.RootDir.Equals(s.ASubdir), "root dir is not .Equal() to its subdirectory")
	assert.False(s.T(), s.ASubdir.Equals(s.RootDir), "root dir is not .Equal() to its subdirectory (opposite order)")
}

func (s *DirectoryTestSuite) TestReversePathLookup() {
	// When reverse lookups are performed on each of the directories
	rootDirPath, rootDirPathLookupErr := s.RootDir.ReversePathLookup()
	aDirPath, aDirPathLookupErr := s.ASubdir.ReversePathLookup()
	bDirPath, bDirPathLookupErr := s.BSubdir.ReversePathLookup()
	cDirPath, cDirPathLookupErr := s.CSubdir.ReversePathLookup()

	// Then expect the correct paths to be returned
	assert.Nil(s.T(), rootDirPathLookupErr, "root path lookup")
	assert.Equal(s.T(), "/", rootDirPath, "root path")

	assert.Nil(s.T(), aDirPathLookupErr, "/a path lookup")
	assert.Equal(s.T(), "/a", aDirPath, "first subdirectory path")

	assert.Nil(s.T(), bDirPathLookupErr, "/a/b path lookup")
	assert.Equal(s.T(), "/a/b", bDirPath, "second subdirectory path")

	assert.Nil(s.T(), cDirPathLookupErr, "/a/b/c path lookup")
	assert.Equal(s.T(), "/a/b/c", cDirPath, "third subdirectory path")
}

func (s *DirectoryTestSuite) TestLookupSelf() {
	rootLookedUpSelf, err := s.RootDir.LookupSubdirectory(directory.SelfDirectoryEntry)
	assert.Nil(s.T(), err, "able to look up self dir entry for root dir")
	assert.True(s.T(), rootLookedUpSelf.Equals(s.RootDir))

	aLookedUpSelf, err := s.ASubdir.LookupSubdirectory(directory.SelfDirectoryEntry)
	assert.Nil(s.T(), err, "able to look up self dir entry for 'a' subdir")
	assert.True(s.T(), aLookedUpSelf.Equals(s.ASubdir))
}

func (s *DirectoryTestSuite) TestLookupParent() {
	rootLookedUp, err := s.ASubdir.LookupSubdirectory(directory.ParentDirectoryEntry)
	assert.Nil(s.T(), err, "able to look up parent (root) directory of 'a' subdir")
	assert.True(s.T(), rootLookedUp.Equals(s.RootDir))

	rootLookedUp, err = s.RootDir.LookupSubdirectory("../../..")
	assert.Nil(s.T(), err, "able to look up parent of root (still root)")
	assert.True(s.T(), s.RootDir.Equals(rootLookedUp))
}

func (s *DirectoryTestSuite) TestLookupConvolutedPath() {
	lookedUp, err := s.BSubdir.LookupSubdirectory("..////////..///a/b/c///////../../b")
	assert.Nil(s.T(), err, "able to look up convoluted (but valid) path")
	assert.True(s.T(), s.BSubdir.Equals(lookedUp))
}

func (s *DirectoryTestSuite) TestLookupInvalidPath() {
	_, err := s.CSubdir.LookupSubdirectory("this/path/does/not/exist")
	assert.NotNil(s.T(), err)
}

func (s *DirectoryTestSuite) TestMkdirFromParent() {
	// Add a `d` subdirectory to `c` from `c`
	dSubdir, err := s.CSubdir.Mkdir("d")
	assert.Nil(s.T(), err)
	dSubdirPath, err := dSubdir.ReversePathLookup()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/a/b/c/d", dSubdirPath)
}

func (s *DirectoryTestSuite) TestMkdirFromGrandparentParent() {
	// Add a `d` subdirectory to `c` from `b`
	dSubdir, err := s.BSubdir.Mkdir("c/d")
	assert.Nil(s.T(), err)
	dSubdirPath, err := dSubdir.ReversePathLookup()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/a/b/c/d", dSubdirPath)
}

func (s *DirectoryTestSuite) TestMkdirFromGrandparentConvolutedParent() {
	// Add a `d` subdirectory to `c` from `b` (via convoluted path)
	dSubdir, err := s.BSubdir.Mkdir("./c/../////c/d")
	assert.Nil(s.T(), err)
	dSubdirPath, err := dSubdir.ReversePathLookup()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "/a/b/c/d", dSubdirPath)
}

func (s *DirectoryTestSuite) TestMkdirInvalidPath() {
	// Attempt adding an `e` subdirectory to non-existent `d` subdir of `c` (from `b`)
	eSubdir, err := s.BSubdir.Mkdir("c/d/e")
	assert.Nil(s.T(), eSubdir)
	assert.NotNil(s.T(), err)
}

func (s *DirectoryTestSuite) TestReadDirOnRoot() {
	entries, err := s.RootDir.ReadDir(directory.SelfDirectoryEntry)
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "a",
			Type: directory.DirectoryType,
		},
		{
			Name: "fizz",
			Type: directory.DirectoryType,
		},
		{
			Name: "buzz",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *DirectoryTestSuite) TestReadDirOnASubdir() {
	entries, err := s.RootDir.ReadDir("a/")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "b",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *DirectoryTestSuite) TestRmdir() {
	// Verify that /a/b has two entries in it
	entries, err := s.BSubdir.ReadDir(directory.SelfDirectoryEntry)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), entries, 2)

	// Remove 'c' subdir from /a/b
	err = s.BSubdir.Rmdir("c")
	assert.Nil(s.T(), err)

	// Verify that /a/b now has one entry for 'foobar'
	entries, err = s.BSubdir.ReadDir(directory.SelfDirectoryEntry)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []directory.DirectoryEntry{{Name: "foobar", Type: directory.DirectoryType}}, entries)

	// Verify behavior of our reference to 'c', since we can still make calls on it
	_, err = s.CSubdir.Mkdir("should_not_be_created")
	assert.NotNil(s.T(), err, "cannot create subdirectories of a deleted directory")
	parentOfC, err := s.CSubdir.LookupSubdirectory(directory.ParentDirectoryEntry)
	assert.Nil(s.T(), err, "can look up parent directory of a deleted directory")
	assert.True(s.T(), parentOfC.Equals(s.BSubdir))
	_, err = s.CSubdir.ReversePathLookup()
	assert.NotNil(s.T(), err, "cannot do reverse path lookup on a deleted directory")
}

func (s *DirectoryTestSuite) TestRmdirNonEmptyDirectory() {
	err := s.ASubdir.Rmdir("b")
	assert.NotNil(s.T(), err, "cannot remove non-empty directory 'b' from /a")
}

func (s *DirectoryTestSuite) TestRmdirSelf() {
	err := s.CSubdir.Rmdir(".")
	assert.NotNil(s.T(), err, "cannot remove self directory entry")
	err = s.CSubdir.Rmdir("..")
	assert.NotNil(s.T(), err, "cannot remove parent directory entry")
}

func TestDirectoryTestSuite(t *testing.T) {
	suite.Run(t, new(DirectoryTestSuite))
}
