package filesys

import (
	"github.com/manderson5192/memfs/directory"
	"github.com/manderson5192/memfs/inode"
)

// FileSystem represents an in-memory filesystem
type FileSystem interface {
	// RootDirectory returns a reference to the filesystem's root directory
	RootDirectory() directory.Directory
}

type fileSystem struct {
	rootDirectory *inode.DirectoryInode
}

// NewFileSystem creates a new FileSystem instance based on an inode tree
func NewFileSystem() FileSystem {
	return &fileSystem{
		rootDirectory: inode.NewRootDirectoryInode(),
	}
}

func (f *fileSystem) RootDirectory() directory.Directory {
	return directory.NewDirectory(f.rootDirectory)
}
