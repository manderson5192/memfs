package inode_test

import (
	"testing"

	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
)

func TestDirectoryInodeType(t *testing.T) {
	i := inode.NewRootDirectoryInode()
	assert.Equal(t, inode.InodeDirectory, i.InodeType())
}

func TestIsRootDirectory(t *testing.T) {
	rootDirectoryInode := inode.NewRootDirectoryInode()
	rootIsRoot := rootDirectoryInode.IsRootDirectoryInode()
	assert.True(t, rootIsRoot, "the root directory's inode should be identified as a root directory inode")

	subdirectoryInode, err := rootDirectoryInode.AddDirectory("subdirectory")
	assert.Nil(t, err, "adding a subdirectory to the root inode should not fail")
	subdirIsRoot := subdirectoryInode.IsRootDirectoryInode()
	assert.False(t, subdirIsRoot, "subdirectory's inode should not be identified as a root directory inode")
}

// TODO: test more of the methods on DirectoryInode
