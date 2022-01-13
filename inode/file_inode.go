package inode

import (
	"io"
	"math"

	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/utils"
	"github.com/pkg/errors"
)

type FileInode struct {
	basicInode
	data []byte
}

func NewFileInode() *FileInode {
	inode := &FileInode{
		data: []byte{},
	}
	return inode
}

func (i *FileInode) InodeType() InodeType {
	return InodeFile
}

func (i *FileInode) Size() int {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	return len(i.data)
}

// ReadAll returns a copy of all of the FileInode's data
func (i *FileInode) ReadAll() []byte {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	toReturn := make([]byte, len(i.data))
	copy(toReturn, i.data)
	return toReturn
}

// TruncateAndWriteAll replaces the FileInode's data with those of d
func (i *FileInode) TruncateAndWriteAll(d []byte) error {
	if d == nil {
		return errors.Wrapf(fserrors.EInval, "buffer is nil")
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	i.data = d
	return nil
}

// ReadAt tries to copy len(p) bytes at offset off from the file into p.  If there are fewer than
// len(p) bytes between the offset and the end of the file, then the error will be non-nil and
// equal to io.EOF.
func (i *FileInode) ReadAt(p []byte, off int64) (int, error) {
	if p == nil {
		return 0, errors.Wrapf(fserrors.EInval, "buffer is nil")
	}
	if off < 0 {
		return 0, errors.Wrapf(fserrors.EInval, "negative offset")
	}
	// Edge case: since `off` is int64 and len(i.data) is `int`, we can only ever read from an offset
	// as large as math.MaxInt
	if off > int64(math.MaxInt) {
		return 0, io.EOF
	}
	intOff := int(off)
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	bytesAfterOffset := utils.Max(len(i.data)-intOff, 0)
	numBytesRequested := len(p)
	numBytesToRead := utils.Min(bytesAfterOffset, numBytesRequested)
	copy(p, i.data[intOff:intOff+numBytesToRead])
	var err error = error(nil)
	// If the number of bytes read is fewer than the number requested, then we need to return EOF
	if numBytesToRead < numBytesRequested {
		// We use io.EOF b/c this error constant is required by the io.ReaderAt interface we are
		// trying to implement
		err = io.EOF
	}
	return numBytesToRead, err
}

// WriteAt attempts copying len(p) bytes from p into the FileInode's data at offset off.  If off is
// beyond the end of the file, then the file is extended with zero bytes up to the offset before
// copying begins.  It returns the number of bytes that were copied, or 0 and an error.
func (i *FileInode) WriteAt(p []byte, off int64) (n int, err error) {
	if p == nil {
		return 0, errors.Wrapf(fserrors.EInval, "buffer is nil")
	}
	if off < 0 {
		return 0, errors.Wrapf(fserrors.EInval, "negative offset")
	}
	// Edge case: since `off` is int64 and len(i.data) is `int`, we can only ever write to an offset
	// as large as math.MaxInt
	if off+int64(len(p)) > int64(math.MaxInt) {
		return 0, errors.Wrapf(fserrors.ENoSpace, "cannot write beyond max file size")
	}
	// Edge case: the above check might pass if off is close to math.MaxInt64, so check for integer
	// wraparound
	if off+int64(len(p)) < 0 {
		return 0, errors.Wrapf(fserrors.ENoSpace, "cannot write beyond max file size")
	}
	intOff := int(off)
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()

	// If (intOff + len(p)) is beyond the end of the file, then we need to pad with zero bytes up to
	// that length
	zeroesToAppend := 0
	if (intOff + len(p)) > len(i.data) {
		zeroesToAppend = intOff + len(p) - len(i.data)
	}
	i.data = append(i.data, make([]byte, zeroesToAppend)...)
	// Do the data copy
	copy(i.data[intOff:intOff+len(p)], p)

	return len(p), nil
}
