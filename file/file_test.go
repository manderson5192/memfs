package file_test

import (
	"testing"

	"github.com/manderson5192/memfs/file"
	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {
	aInode := inode.NewFileInode()
	aFile := file.NewFile(aInode)
	aOtherFile := file.NewFile(aInode)

	bInode := inode.NewFileInode()
	bFile := file.NewFile(bInode)

	assert.True(t, aFile.Equals(aFile), "file is equal to itself")
	assert.True(t, aFile.Equals(aOtherFile), "a file is equal to another file ref'ing the same inode")
	assert.True(t, aOtherFile.Equals(aFile), "file equality is symmetric")
	assert.False(t, aFile.Equals(bFile), "two files with different inodes are not equal")
	assert.False(t, bFile.Equals(aFile), "file inequality is symmetric")
}
