package file_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/filesys"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
	Filesys filesys.FileSystem
	RootDir directory.Directory
	File    file.File
}

func (s *FileTestSuite) SetupTest() {
	// Create a simple filesystem with a file "/file"
	s.Filesys = filesys.NewFileSystem()
	s.RootDir = s.Filesys.RootDirectory()
	f, err := s.RootDir.CreateFile("file")
	s.File = f
	assert.Nil(s.T(), err)
}

func (s *FileTestSuite) TestEquals() {
	aInode := inode.NewFileInode()
	aFile := file.NewFile(aInode)
	aOtherFile := file.NewFile(aInode)

	bInode := inode.NewFileInode()
	bFile := file.NewFile(bInode)

	assert.True(s.T(), aFile.Equals(aFile), "file is equal to itself")
	assert.True(s.T(), aFile.Equals(aOtherFile), "a file is equal to another file ref'ing the same inode")
	assert.True(s.T(), aOtherFile.Equals(aFile), "file equality is symmetric")
	assert.False(s.T(), aFile.Equals(bFile), "two files with different inodes are not equal")
	assert.False(s.T(), bFile.Equals(aFile), "file inequality is symmetric")
}

// This test doesn't assert any functional behavior so much as it asserts that the File interface
// implements the following Go io package interfaces:
// * Reader
// * Writer
// * ReaderAt
// * WriterAt
// * Seeker
func (s *FileTestSuite) TestImplementsInterfaces() {
	file := file.NewFile(inode.NewFileInode())
	var _ io.Reader = file
	var _ io.Writer = file
	var _ io.ReaderAt = file
	var _ io.WriterAt = file
	var _ io.Seeker = file
}

func (s *FileTestSuite) TestRead() {
	err := s.File.TruncateAndWriteAll([]byte("hello, world!"))
	assert.Nil(s.T(), err)
	data := make([]byte, 0, s.File.Size())
	for {
		buf := make([]byte, 1)
		n, err := s.File.Read(buf)
		if err != nil {
			if err == io.EOF {
				data = append(data, buf[:n]...)
				break
			}
			assert.FailNow(s.T(), "Read() returned non-nil error", err)
		}
		data = append(data, buf[:n]...)
	}
	assert.Equal(s.T(), "hello, world!", string(data))
}

func (s *FileTestSuite) TestWrite() {
	// Write all the bytes in "hello"
	for _, ch := range "hello" {
		n, err := s.File.Write([]byte{byte(ch)})
		if n != 1 {
			assert.FailNow(s.T(), "Write() returned value other than 1")
		}
		if err != nil {
			assert.FailNow(s.T(), "Write() returned non-nil error", err)
		}
	}

	// Read() should produce nothing, since we're at EOF
	buf := make([]byte, len("hello"))
	n, err := s.File.Read(buf)
	assert.Equal(s.T(), 0, n)
	assert.Equal(s.T(), io.EOF, err)

	// Reading all of file's contents should produce "hello"
	data := s.File.ReadAll()
	assert.Equal(s.T(), "hello", string(data))
}

func (s *FileTestSuite) TestSeek() {
	// Seed the file with some data
	err := s.File.TruncateAndWriteAll([]byte("hello"))
	assert.Nil(s.T(), err)

	// Seek to offset 1000 from the beginning
	offset, err := s.File.Seek(1000, io.SeekStart)
	assert.Equal(s.T(), int64(1000), offset)
	assert.Nil(s.T(), err)

	// Seek to offset 500 from the current offset
	offset, err = s.File.Seek(-500, io.SeekCurrent)
	assert.Equal(s.T(), int64(500), offset)
	assert.Nil(s.T(), err)

	// Seek to offset 7 from the end of the file
	offset, err = s.File.Seek(2, io.SeekEnd)
	assert.Equal(s.T(), int64(7), offset)
	assert.Nil(s.T(), err)

	// Seek to an illegal offset from the current offset
	offset, err = s.File.Seek(-10, io.SeekCurrent)
	assert.Equal(s.T(), int64(7), offset, "offset is unchanged from failed Seek() call")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)

	// Check the file size -- should be the original size
	assert.Equal(s.T(), len("hello"), s.File.Size())
}

func (s *FileTestSuite) TestIOUtilReadAll() {
	// Seed the file with some data
	err := s.File.TruncateAndWriteAll([]byte("Lorem ipsum dolor sit amet."))
	assert.Nil(s.T(), err)

	// Check that ioutil.ReadAll() works with File's io.Reader implementation
	data, err := ioutil.ReadAll(s.File)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Lorem ipsum dolor sit amet.", string(data))
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}
