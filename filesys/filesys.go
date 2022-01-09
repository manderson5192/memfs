package filesys

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/inode"
)

type FileSystem interface {
	RootDirectory() directory.Directory
}

type fileSystem struct {
	rootDirectory *inode.DirectoryInode
}

func NewFileSystem() FileSystem {
	return &fileSystem{
		rootDirectory: inode.NewRootDirectoryInode(),
	}
}

func (f *fileSystem) RootDirectory() directory.Directory {
	return directory.NewDirectory(f.rootDirectory)
}
