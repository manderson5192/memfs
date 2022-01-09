package utils_test

import (
	"testing"

	"github.com/manderson5192/memfs/utils"
	"github.com/stretchr/testify/assert"
)

func TestCut(t *testing.T) {
	// Given a separator and two strings w/ and w/o the separator
	separator := "|"
	containsSeparator := "foo|bar"
	doesNotContainSeparator := "foobar"

	// When Cut() is called
	containsBefore, containsAfter, containsFound := utils.Cut(containsSeparator, separator)
	doesNotContainBefore, doesNotContainAfter, doesNotContainFound := utils.Cut(doesNotContainSeparator, separator)

	// Then
	assert.Equal(t, "foo", containsBefore)
	assert.Equal(t, "bar", containsAfter)
	assert.True(t, containsFound)
	assert.Equal(t, "foobar", doesNotContainBefore)
	assert.Equal(t, "", doesNotContainAfter)
	assert.False(t, doesNotContainFound)
}

func TestRightCut(t *testing.T) {
	// Given a separator and two strings w/ and w/o the separator
	separator := "|"
	containsSeparator := "foo|bar|fizzbuzz"
	doesNotContainSeparator := "foobar"

	// When RightCut() is called
	containsBefore, containsAfter, containsFound := utils.RightCut(containsSeparator, separator)
	doesNotContainBefore, doesNotContainAfter, doesNotContainFound := utils.RightCut(doesNotContainSeparator, separator)

	// Then
	assert.Equal(t, "foo|bar", containsBefore)
	assert.Equal(t, "fizzbuzz", containsAfter)
	assert.True(t, containsFound)
	assert.Equal(t, "", doesNotContainBefore)
	assert.Equal(t, "foobar", doesNotContainAfter)
	assert.False(t, doesNotContainFound)
}
