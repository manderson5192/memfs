package inode_test

import (
	"testing"

	"github.com/manderson5192/memfs/inode"
	"github.com/stretchr/testify/assert"
)

func TestInodeTypeString(t *testing.T) {
	assert.Equal(t, "InodeFile", inode.InodeFile.String())
	assert.Equal(t, "InodeDirectory", inode.InodeDirectory.String())
	assert.Equal(t, "InodeInvalid", inode.InodeInvalid.String())
	assert.Equal(t, "InodeInvalid", inode.InodeType(42).String())
}
