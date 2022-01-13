package process_test

import (
	"io"

	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/modes"
	"github.com/stretchr/testify/assert"
)

func (s *ProcessTestSuite) TestCreateFileWithTrailingSlash() {
	_, err := s.p.CreateFile("/filename/")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *ProcessTestSuite) TestOpenFileWithTrailingSlash() {
	_, err := s.p.OpenFile("/a/foobar_file/", modes.CombineModes(modes.O_RDWR))
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

func (s *ProcessTestSuite) TestOpenFileReadOnly() {
	f, err := s.p.OpenFile("/a/foobar_file", modes.O_RDONLY|modes.O_CREATE)
	assert.Nil(s.T(), err)

	// Try writing to the file a few different ways -- all should fail with EINVAL
	toWrite := []byte("some data")
	assertContents := func() {
		contents, err := f.ReadAll()
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), []byte("hello!"), contents)
	}

	err = f.TruncateAndWriteAll(toWrite)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assertContents()

	n, err := f.WriteAt(toWrite, 0)
	assert.Equal(s.T(), 0, n)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assertContents()

	n, err = f.Write(toWrite)
	assert.Equal(s.T(), 0, n)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assertContents()
}

func (s *ProcessTestSuite) TestOpenFileWriteOnly() {
	f, err := s.p.OpenFile("/a/foobar_file", modes.O_WRONLY)
	assert.Nil(s.T(), err)

	// Verify that reading is not possible.  It should result in EINVAL.
	_, err = f.ReadAll()
	assert.ErrorIs(s.T(), err, fserrors.EInval)

	buf := make([]byte, len("hello!"))
	_, err = f.ReadAt(buf, 0)
	assert.ErrorIs(s.T(), err, fserrors.EInval)

	buf = make([]byte, len("hello!"))
	_, err = f.Read(buf)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
}

func (s *ProcessTestSuite) TestOpenFileAppend() {
	f, err := s.p.OpenFile("/a/foobar_file", modes.O_RDWR|modes.O_APPEND)
	assert.Nil(s.T(), err)

	// Verify that WriteAt() and TruncateAndWriteAll() result in EINVAL
	toWrite := []byte(" some data")
	assertContents := func(expected string) {
		contents, err := f.ReadAll()
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), expected, string(contents))
	}
	err = f.TruncateAndWriteAll(toWrite)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assertContents("hello!")
	_, err = f.WriteAt(toWrite, 0)
	assert.ErrorIs(s.T(), err, fserrors.EInval)
	assertContents("hello!")

	// Now do a regular Write()
	n, err := f.Write(toWrite)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(toWrite), n)

	// Check out the Seek()
	offset, err := f.Seek(0, io.SeekCurrent)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(len("hello!")+len(toWrite)), offset)

	// Now verify the contents
	assertContents("hello! some data")
}

func (s *ProcessTestSuite) TestOpenFileTruncate() {
	f, err := s.p.OpenFile("/a/foobar_file", modes.O_RDWR|modes.O_TRUNC)
	assert.Nil(s.T(), err)
	data, err := f.ReadAll()
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), data)
}

func (s *ProcessTestSuite) TestOpenFileCreateFileDNE() {
	_, err := s.p.OpenFile("/a/does_not_exist.txt", modes.O_RDWR|modes.O_CREATE)
	assert.Nil(s.T(), err)
}

func (s *ProcessTestSuite) TestOpenFileCreateFileExists() {
	_, err := s.p.OpenFile("/a/foobar_file", modes.O_RDWR|modes.O_CREATE)
	assert.Nil(s.T(), err)
}

func (s *ProcessTestSuite) TestOpenFileCreateExclusiveFileDNE() {
	_, err := s.p.OpenFile("/a/does_not_exist.txt", modes.O_RDWR|modes.O_CREATE|modes.O_EXCL)
	assert.Nil(s.T(), err)
}

func (s *ProcessTestSuite) TestOpenFileCreateExclusiveFileExists() {
	_, err := s.p.OpenFile("/a/foobar_file", modes.O_RDWR|modes.O_CREATE|modes.O_EXCL)
	assert.ErrorIs(s.T(), err, fserrors.EExist)
}
