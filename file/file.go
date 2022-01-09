package file

import "github.com/manderson5192/memfs/inode"

type File interface {
	Equals(other File) bool
}

type file struct {
	*inode.FileInode
}

func NewFile(inode *inode.FileInode) File {
	return &file{
		inode,
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
