package inode

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
