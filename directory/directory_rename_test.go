package directory_test

import (
	"testing"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DirectoryRenameTestSuite struct {
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

func (s *DirectoryRenameTestSuite) SetupTest() {
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

func (s *DirectoryRenameTestSuite) TestRenameSameDirectory() {
	// Verify the entries in /a/b before renaming c
	entries, err := s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Add a file to c
	fileInC, err := s.CSubdir.CreateFile("a_file")
	assert.Nil(s.T(), err)

	// Do the rename
	err = s.RootDir.Rename("a/b/c", "a/b/c_newname")
	assert.Nil(s.T(), err)

	// Verify the entries after renaming c
	entries, err = s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c_newname",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Make sure that the file in c is available under c_newname
	fileInCNewName, err := s.RootDir.OpenFile("a/b/c_newname/a_file")
	assert.Nil(s.T(), err)
	assert.True(s.T(), fileInCNewName.Equals(fileInC))
}

func (s *DirectoryRenameTestSuite) TestRenameOverEmptyDirSameDirectory() {
	// Verify the entries in /a/b before renaming c
	entries, err := s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Add a file to c
	fileInC, err := s.CSubdir.CreateFile("a_file")
	assert.Nil(s.T(), err)

	// Do the rename
	err = s.RootDir.Rename("a/b/c", "a/b/foobar")
	assert.Nil(s.T(), err)

	// Verify the entries after renaming c
	entries, err = s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Make sure that the file in c is available under c_newname
	fileInCNewName, err := s.RootDir.OpenFile("a/b/foobar/a_file")
	assert.Nil(s.T(), err)
	assert.True(s.T(), fileInCNewName.Equals(fileInC))
}

func (s *DirectoryRenameTestSuite) TestRenameOverNonemptyDirSameDirectory() {
	// Verify the entries in /a/b before renaming c
	entries, err := s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Add a file to foobar
	_, err = s.BSubdir.CreateFile("foobar/a_file")
	assert.Nil(s.T(), err)

	// Do the rename
	err = s.RootDir.Rename("a/b/c", "a/b/foobar")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENotEmpty)

	// Verify the entries after attempted renaming
	entries, err = s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
	}, entries)
}

func (s *DirectoryRenameTestSuite) TestRenameOverFileSameDirectory() {
	// Add a file to /a/b
	_, err := s.BSubdir.CreateFile("some_file")
	assert.Nil(s.T(), err)

	// Verify the entries in /a/b before renaming c
	entries, err := s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "c",
			Type: directory.DirectoryType,
		},
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
		{
			Name: "some_file",
			Type: directory.FileType,
		},
	}, entries)

	// Add a file to c
	fileInC, err := s.CSubdir.CreateFile("a_file")
	assert.Nil(s.T(), err)

	// Do the rename
	err = s.RootDir.Rename("a/b/c", "a/b/some_file")
	assert.Nil(s.T(), err)

	// Verify the entries after renaming c
	entries, err = s.RootDir.ReadDir("a/b")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "foobar",
			Type: directory.DirectoryType,
		},
		{
			Name: "some_file",
			Type: directory.DirectoryType,
		},
	}, entries)

	// Make sure that the file in c is available under some_file
	fileInCNewName, err := s.RootDir.OpenFile("a/b/some_file/a_file")
	assert.Nil(s.T(), err)
	assert.True(s.T(), fileInCNewName.Equals(fileInC))

	// Verify that /a/b/some_file was deleted
	_, err = s.BSubdir.OpenFile("some_file")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EIsDir)
}

func (s *DirectoryRenameTestSuite) TestRenameFile() {
	// Create a file in 'c'
	someFile, err := s.CSubdir.CreateFile("some_file")
	assert.Nil(s.T(), err)

	// Verify that some_file is in 'c'
	entries, err := s.CSubdir.ReadDir(".")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "some_file",
			Type: directory.FileType,
		},
	}, entries)

	// Move some_file to the root directory
	err = s.RootDir.Rename("a/b/c/some_file", "./some_file")
	assert.Nil(s.T(), err)

	// Verify that some_file is not in 'c'
	entries, err = s.CSubdir.ReadDir(".")
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), entries)

	// Verify that some_file is in the root directory now
	someFileInRoot, err := s.RootDir.OpenFile("some_file")
	assert.Nil(s.T(), err)
	assert.True(s.T(), someFile.Equals(someFileInRoot))
}

func (s *DirectoryRenameTestSuite) TestRenameDirectory() {
	// Verify entries in /
	entries, err := s.RootDir.ReadDir(".")
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

	// Create a file in 'c'
	someFile, err := s.CSubdir.CreateFile("some_file")

	// Verify that some_file is in 'c'
	entries, err = s.CSubdir.ReadDir(".")
	assert.Nil(s.T(), err)
	assert.ElementsMatch(s.T(), []directory.DirectoryEntry{
		{
			Name: "some_file",
			Type: directory.FileType,
		},
	}, entries)
	assert.Nil(s.T(), err)

	// Move /a/b/c to /c
	err = s.BSubdir.Rename("../b/./c", "../../c")

	// Verify that 'c' is not in 'b' any more
	entries, err = s.BSubdir.ReadDir(".")
	assert.Nil(s.T(), err)
	assert.NotContains(s.T(), entries, directory.DirectoryEntry{Name: "c", Type: directory.DirectoryType})

	// Verify that 'c' is in '/' now
	entries, err = s.RootDir.ReadDir(".")
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), entries, directory.DirectoryEntry{Name: "c", Type: directory.DirectoryType})

	// Verify that some_file is under /c now
	someFileInRoot, err := s.RootDir.OpenFile("./c/some_file")
	assert.Nil(s.T(), err)
	assert.True(s.T(), someFile.Equals(someFileInRoot))
}

func (s *DirectoryRenameTestSuite) TestRenameFromSpecialSelfDirectory() {
	err := s.ASubdir.Rename(".", "new_self")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *DirectoryRenameTestSuite) TestRenameFromSpecialParentDirectory() {
	err := s.ASubdir.Rename("..", "new_parent")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *DirectoryRenameTestSuite) TestRenameOverSpecialSelfDirectory() {
	err := s.ASubdir.Rename("b", "b/c/..")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *DirectoryRenameTestSuite) TestRenameOverSpecialParentDirectory() {
	err := s.ASubdir.Rename("b", "b/c/..")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func TestDirectoryRenameTestSuite(t *testing.T) {
	suite.Run(t, new(DirectoryRenameTestSuite))
}
