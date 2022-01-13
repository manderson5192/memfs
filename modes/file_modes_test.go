package modes_test

import (
	"testing"

	"github.com/manderson5192/memfs/modes"
	"github.com/stretchr/testify/assert"
)

func TestIsReadOnly(t *testing.T) {
	assert.True(t, modes.IsReadOnly(0))
	assert.True(t, modes.IsReadOnly(modes.O_RDONLY))
	assert.True(t, modes.IsReadOnly(modes.CombineModes(modes.O_RDONLY)))
	assert.True(t, modes.IsReadOnly(modes.CombineModes(modes.O_RDONLY, modes.O_CREATE)))
	assert.True(t, modes.IsReadOnly(modes.CombineModes(modes.O_RDONLY, modes.O_CREATE, modes.O_EXCL)))
	assert.False(t, modes.IsReadOnly(modes.O_WRONLY))
	assert.False(t, modes.IsReadOnly(modes.CombineModes(modes.O_WRONLY)))
	assert.False(t, modes.IsReadOnly(modes.O_RDWR))
	assert.False(t, modes.IsReadOnly(modes.CombineModes(modes.O_RDWR)))
}

func TestIsWriteAllowed(t *testing.T) {
	assert.False(t, modes.IsWriteAllowed(0))
	assert.False(t, modes.IsWriteAllowed(modes.O_RDONLY))
	assert.True(t, modes.IsWriteAllowed(modes.O_WRONLY))
	assert.True(t, modes.IsWriteAllowed(modes.O_RDWR))
	assert.True(t, modes.IsWriteAllowed(modes.CombineModes(modes.O_WRONLY, modes.O_APPEND, modes.O_CREATE)))
	assert.True(t, modes.IsWriteAllowed(modes.CombineModes(modes.O_RDWR, modes.O_TRUNC)))
}

func TestIsCreateMode(t *testing.T) {
	assert.False(t, modes.IsCreateMode(0))
	assert.True(t, modes.IsCreateMode(modes.O_CREATE))
	assert.True(t, modes.IsCreateMode(modes.CombineModes(modes.O_CREATE, modes.O_EXCL)))
}

func TestIsExclMode(t *testing.T) {
	assert.False(t, modes.IsExclusiveMode(0))
	assert.False(t, modes.IsExclusiveMode(modes.O_EXCL))
	assert.True(t, modes.IsExclusiveMode(modes.O_CREATE|modes.O_EXCL))
}
