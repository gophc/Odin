package gochibicc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharPtr(t *testing.T) {
	buf := []byte{
		'a', 'b', 'c', 'd', '\x00', 'c', 'e', 'f',
	}
	p1 := Bytes2CharPtr(buf)
	p2 := String2CharPtr(string(buf))
	p3 := String2CharPtr(Bytes2String(buf))

	assert.Equal(t, p1.str(), p2.str())
	assert.Equal(t, p2.str(), p3.str())
	assert.Equal(t, p1.str(), "abcd")

	assert.Equal(t, p1.xi(), byte('a'))
	assert.Equal(t, p1.xipp(), byte('a'))
	assert.Equal(t, p1.xppi(), byte('c'))
	assert.Equal(t, p1.xiss(), byte('c'))
	assert.Equal(t, p1.xssi(), byte('a'))

	assert.Equal(t, p2.xi(), p3.xi())
	assert.Equal(t, p2.xipp(), p3.xipp())
	assert.Equal(t, p2.xppi(), p3.xppi())
	assert.Equal(t, p2.xiss(), p3.xiss())
	assert.Equal(t, p2.xssi(), p3.xssi())
}
