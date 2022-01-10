package inode_test

import (
	"io"
	"testing"

	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FileInodeTestSuite struct {
	suite.Suite
	FileInode *inode.FileInode
}

func (s *FileInodeTestSuite) SetupTest() {
	s.FileInode = inode.NewFileInode()
}

func (s *FileInodeTestSuite) TestFileInodeType() {
	assert.Equal(s.T(), inode.InodeFile, s.FileInode.InodeType())
}

// This test doesn't verify any functionality.  Instead, it asserts that FileInode implements the
// contract of Go's io.ReaderAt and io.WriterAt interfaces.
func (s *FileInodeTestSuite) TestFileInodeImplementsInterfaces() {
	var _ io.ReaderAt = s.FileInode
	var _ io.WriterAt = s.FileInode
}

func (s *FileInodeTestSuite) TestReadAndWriteAll() {
	// Read empty file
	buf := s.FileInode.ReadAll()
	assert.Empty(s.T(), buf)

	// Write some data
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)

	// Read all the data
	buf = s.FileInode.ReadAll()
	assert.Equal(s.T(), "hello, world!", string(buf))
}

func (s *FileInodeTestSuite) TestTruncateAndWriteAllWithNil() {
	err := s.FileInode.TruncateAndWriteAll(nil)
	assert.NotNil(s.T(), err)
}

func (s *FileInodeTestSuite) TestReadAtEmptyFile() {
	// ReadAt on empty file
	buf := make([]byte, 5)
	n, err := s.FileInode.ReadAt(buf, 0)
	assert.Zero(s.T(), n)
	assert.Equal(s.T(), io.EOF, err)
}

func (s *FileInodeTestSuite) TestReadAtPartOfFile() {
	// Add content to file
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)

	// ReadAt on non-empty file
	buf := make([]byte, 5)
	n, err := s.FileInode.ReadAt(buf, 0)
	assert.Equal(s.T(), 5, n)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "hello", string(buf))
}

func (s *FileInodeTestSuite) TestReadAtWholeFile() {
	// Add content to file
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)

	// ReadAt all of the data
	fullSizeBuf := make([]byte, len("hello, world!"))
	n, err := s.FileInode.ReadAt(fullSizeBuf, 0)
	assert.Equal(s.T(), len("hello, world!"), n)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "hello, world!", string(fullSizeBuf))
}

func (s *FileInodeTestSuite) TestReadAtWholeFilePlusOne() {
	// Add content to file
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)

	// ReadAt all of the data, plus one more byte than is available
	fullSizeBufPlusOne := make([]byte, len("hello, world!")+1)
	n, err := s.FileInode.ReadAt(fullSizeBufPlusOne, 0)
	assert.Equal(s.T(), len("hello, world!"), n)
	assert.Equal(s.T(), io.EOF, err)
	assert.Equal(s.T(), append([]byte("hello, world!"), 0), fullSizeBufPlusOne)
}

func (s *FileInodeTestSuite) TestReadAtPartwayThroughFile() {
	// Add content to file
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)

	// ReadAt partway through the file
	buf := make([]byte, len("world"))
	n, err := s.FileInode.ReadAt(buf, int64(len("hello, ")))
	assert.Equal(s.T(), len("world"), n)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "world", string(buf))
}

func (s *FileInodeTestSuite) TestReadAtNil() {
	n, err := s.FileInode.ReadAt(nil, 0)
	assert.Zero(s.T(), n)
	assert.NotNil(s.T(), err)
}

func (s *FileInodeTestSuite) TestReadAtNegativeOffset() {
	n, err := s.FileInode.ReadAt(nil, -100)
	assert.Zero(s.T(), n)
	assert.NotNil(s.T(), err)
}

func (s *FileInodeTestSuite) TestWriteAtBeginningOfEmptyFile() {
	n, err := s.FileInode.WriteAt([]byte("hello, world!"), 0)
	assert.Equal(s.T(), len("hello, world!"), n)
	assert.Nil(s.T(), err)
	data := s.FileInode.ReadAll()
	assert.Equal(s.T(), "hello, world!", string(data))
}

func (s *FileInodeTestSuite) TestWriteAtPastBeginningOfEmptyFile() {
	n, err := s.FileInode.WriteAt([]byte("hello, world!"), 4)
	assert.Equal(s.T(), len("hello, world!"), n)
	assert.Nil(s.T(), err)
	data := s.FileInode.ReadAll()
	assert.Equal(s.T(), append([]byte{0, 0, 0, 0}, []byte("hello, world!")...), data)
}

func (s *FileInodeTestSuite) TestWriteAtOverwrite() {
	err := s.FileInode.TruncateAndWriteAll([]byte("hello, world"))
	assert.Nil(s.T(), err)
	n, err := s.FileInode.WriteAt([]byte("nobody"), int64(len("hello, ")))
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len("nobody"), n)
	data := s.FileInode.ReadAll()
	assert.Equal(s.T(), "hello, nobody", string(data))
}

func TestFileInodeTestSuite(t *testing.T) {
	suite.Run(t, new(FileInodeTestSuite))
}
