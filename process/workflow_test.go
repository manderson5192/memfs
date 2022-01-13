package process_test

import (
	"io"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/manderson5192/memfs/filesys"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/os"
	"github.com/manderson5192/memfs/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// WorkflowTestSuite is a test suite designed to validate certain filesystem behaviors that only
// happen when multiple system calls are made together
type WorkflowTestSuite struct {
	suite.Suite
	fs filesys.FileSystem
	p  process.ProcessFilesystemContext
}

func (s *WorkflowTestSuite) SetupTest() {
	// Setup a process context and a basic file tree
	s.fs = filesys.NewFileSystem()
	s.p = process.NewProcessFilesystemContext(s.fs)
	assert.Nil(s.T(), s.p.MakeDirectory("/a"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/zzz"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b/c"))
	assert.Nil(s.T(), s.p.MakeDirectory("/a/b/a"))
	foobarFile, err := s.p.CreateFile("/a/foobar_file")
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), foobarFile.TruncateAndWriteAll([]byte("hello!")))
}

func (s *WorkflowTestSuite) TestFileAccessWorksAfterDeletion() {
	// Open foobar_file
	f, err := s.p.OpenFile("/a/foobar_file", os.CombineModes(os.O_RDWR))
	assert.Nil(s.T(), err)

	// Delete foobar_file
	err = s.p.DeleteFile("/a/foobar_file")
	assert.Nil(s.T(), err)

	// Foobar file cannot be found
	_, err = s.p.Stat("/a/foobar_file")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENoEnt)

	// Read the file's contents
	data, err := ioutil.ReadAll(f)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "hello!", string(data))

	// Overwrite the file
	err = f.TruncateAndWriteAll([]byte("hello, world"))
	assert.Nil(s.T(), err)

	// Seek to the file's end
	_, err = f.Seek(0, io.SeekEnd)
	assert.Nil(s.T(), err)

	// Write an exclamation point
	_, err = f.Write([]byte("!"))
	assert.Nil(s.T(), err)

	// Read back all of the file's contents
	data, err = f.ReadAll()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "hello, world!", string(data))

	// Re-verify that the file cannot be found in the filesystem
	_, err = s.p.Stat("/a/b/../foobar_file")
	assert.NotNil(s.T(), err)
	assert.ErrorIs(s.T(), err, fserrors.ENoEnt)
}

func (s *WorkflowTestSuite) TestFileAccessWorksThroughRename() {
	// Open foobar_file
	f1, err := s.p.OpenFile("/a/foobar_file", os.CombineModes(os.O_RDWR))
	assert.Nil(s.T(), err)

	// Move foobar_file
	err = s.p.Rename("/a/foobar_file", "/a/b/foobar_file")
	assert.Nil(s.T(), err)

	// Reopen foobar_file in its new location
	f2, err := s.p.OpenFile("/a/b/foobar_file", os.CombineModes(os.O_RDWR))
	assert.Nil(s.T(), err)

	// Both f1 and f2 contain the initial contents
	data1, err := f1.ReadAll()
	assert.Nil(s.T(), err)
	data2, err := f2.ReadAll()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), string(data1), string(data2))

	// Writing to f1 results in visible changes at f2
	err = f1.TruncateAndWriteAll([]byte("new content"))
	assert.Nil(s.T(), err)
	data2, err = f2.ReadAll()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "new content", string(data2))
}

func (s *WorkflowTestSuite) TestManyConcurrentFileAccesses() {
	var wg sync.WaitGroup
	for offset, ch := range "abcdefghijklmnopqrstuvwxyz" {
		wg.Add(1)
		go func(o int, r rune) {
			// Open foobar_file
			f, err := s.p.OpenFile("/a/foobar_file", os.CombineModes(os.O_RDWR))
			assert.Nil(s.T(), err)

			// Sleep for a random number of milliseconds between 0 and 100
			ms := rand.Intn(100)
			time.Sleep(time.Millisecond * time.Duration(ms))

			// Write rune r at offset o
			n, err := f.WriteAt([]byte(string(r)), int64(o))
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), 1, n)

			// Notify the WaitGroup that we are done
			wg.Done()
		}(offset, ch)
	}
	// Wait for all goroutines to finish
	wg.Wait()

	// Verify that the resultant file is what we expect
	f, err := s.p.OpenFile("/a/foobar_file", os.CombineModes(os.O_RDWR))
	assert.Nil(s.T(), err)
	data, err := f.ReadAll()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "abcdefghijklmnopqrstuvwxyz", string(data))
}

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
