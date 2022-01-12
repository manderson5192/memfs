package file

import (
	"io"
	"sync"

	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/inode"
	"github.com/pkg/errors"
)

// File is a typical file abstraction, representing a file descriptor and an offset.  To hold a
// file open is to hold a reference to a non-nil File.  To close it is to let the garbage collector
// do its work by losing any reference to this File.  Access to this File's offset is synchronized
// on a per-file basis, but operations to the underlying file data are synchronized at the inode
// layer.
type File interface {
	// Equals returns true if the other file is backed by the same FileInode
	Equals(other File) bool
	// ReadAll returns a copy of all of the data in the file.  It does not affect the file offset.
	ReadAll() []byte
	// TruncateAndWriteAll truncates the file and writes in all of the data in buf.  It returns an
	// error on failure.  It does not affect the file offset
	TruncateAndWriteAll(buf []byte) error
	// ReadAt tries to copy len(p) bytes at offset off from the file into p.  If there are fewer than
	// len(p) bytes between the offset and the end of the file, then the error will be non-nil and
	// equal to io.EOF.
	ReadAt(p []byte, off int64) (int, error)
	// WriteAt attempts copying len(p) bytes from p into the FileInode's data at offset off.  If off is
	// beyond the end of the file, then the file is extended with zero bytes up to the offset before
	// copying begins.  It returns the number of bytes that were copied, or 0 and an error.
	WriteAt(p []byte, off int64) (int, error)
	// Size returns the size of the file in bytes
	Size() int
	io.Reader
	io.Writer
	io.Seeker
}

type file struct {
	*inode.FileInode
	offset int64
	mutex  sync.Mutex // synchronizes access to this file's offset
}

func NewFile(inode *inode.FileInode) File {
	return &file{
		FileInode: inode,
		offset:    0,
	}
}

func (f *file) Equals(other File) bool {
	if f == nil || other == nil {
		return false
	}
	otherFile, ok := other.(*file)
	if !ok {
		return false
	}
	return f.FileInode == otherFile.FileInode
}

func (f *file) Read(p []byte) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	n, err := f.ReadAt(p, f.offset)
	f.offset += int64(n)
	return n, err
}

func (f *file) Write(p []byte) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	n, err := f.WriteAt(p, f.offset)
	f.offset += int64(n)
	return n, err
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// interpret whence
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset = f.offset + offset
	case io.SeekEnd:
		offset = int64(f.Size()) + offset
	default:
		return f.offset, errors.Wrapf(fserrors.EInval, "invalid whence value %d", whence)
	}
	// check if the resultant offset is valid
	if offset < 0 {
		return f.offset, errors.Wrapf(fserrors.EInval, "negative offset")
	}
	f.offset = offset
	return f.offset, nil
}
