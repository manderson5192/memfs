package inode

import "sync"

// InodeType is an enum that indicates whether an inode is a file or a directory
type InodeType int

const (
	InodeInvalid InodeType = iota
	InodeFile
	InodeDirectory
)

// Inode represents a filesystem inode ("index node") and is implemented by one of two types:
// *DirectoryInode and *FileInode
type Inode interface {
	InodeType() InodeType
	// Size will return the number of bytes in a FileInode's data buffer or the number of entries
	// in a DirectoryInode's entry table
	Size() int
}

type basicInode struct {
	rwMutex sync.RWMutex
}

func (i InodeType) String() string {
	if i == InodeFile {
		return "InodeFile"
	} else if i == InodeDirectory {
		return "InodeDirectory"
	} else {
		return "InodeInvalid"
	}
}

func IsDirectory(i Inode) bool {
	return i.InodeType() == InodeDirectory
}

func IsFile(i Inode) bool {
	return i.InodeType() == InodeFile
}
