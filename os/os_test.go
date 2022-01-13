package os_test

import (
	"testing"

	"github.com/manderson5192/memfs/os"
	"github.com/stretchr/testify/assert"
)

func TestIsReadOnly(t *testing.T) {
	assert.True(t, os.IsReadOnly(0))
	assert.True(t, os.IsReadOnly(os.O_RDONLY))
	assert.True(t, os.IsReadOnly(os.CombineModes(os.O_RDONLY)))
	assert.True(t, os.IsReadOnly(os.CombineModes(os.O_RDONLY, os.O_CREATE)))
	assert.True(t, os.IsReadOnly(os.CombineModes(os.O_RDONLY, os.O_CREATE, os.O_EXCL)))
	assert.False(t, os.IsReadOnly(os.O_WRONLY))
	assert.False(t, os.IsReadOnly(os.CombineModes(os.O_WRONLY)))
	assert.False(t, os.IsReadOnly(os.O_RDWR))
	assert.False(t, os.IsReadOnly(os.CombineModes(os.O_RDWR)))
}

func TestIsWriteAllowed(t *testing.T) {
	assert.False(t, os.IsWriteAllowed(0))
	assert.False(t, os.IsWriteAllowed(os.O_RDONLY))
	assert.True(t, os.IsWriteAllowed(os.O_WRONLY))
	assert.True(t, os.IsWriteAllowed(os.O_RDWR))
	assert.True(t, os.IsWriteAllowed(os.CombineModes(os.O_WRONLY, os.O_APPEND, os.O_CREATE)))
	assert.True(t, os.IsWriteAllowed(os.CombineModes(os.O_RDWR, os.O_TRUNC)))
}

func TestIsCreateMode(t *testing.T) {
	assert.False(t, os.IsCreateMode(0))
	assert.True(t, os.IsCreateMode(os.O_CREATE))
	assert.True(t, os.IsCreateMode(os.CombineModes(os.O_CREATE, os.O_EXCL)))
}

func TestIsExclMode(t *testing.T) {
	assert.False(t, os.IsExclusiveMode(0))
	assert.False(t, os.IsExclusiveMode(os.O_EXCL))
	assert.True(t, os.IsExclusiveMode(os.O_CREATE|os.O_EXCL))
}
