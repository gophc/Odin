package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//goland:noinspection GoBoolExpressions
func TestCompile(t *testing.T) {
	t.Run("TestCompile", func(t *testing.T) {
		assert.Equal(t, 1, 1, "Some Error.")
	})
}
