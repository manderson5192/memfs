package inode

import "sync"

type InodeType int

const (
	InodeInvalid InodeType = iota
	InodeFile
	InodeDirectory
)

type Inode interface {
	InodeType() InodeType
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
