package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	header := NewHeader()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := header.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, header)
	assert.Equal(t, "localhost:42069", header["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	header = NewHeader()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = header.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	header = NewHeader()
	data = []byte("Host:   localhost:42069\r\n\r\n")
	n, done, err = header.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, header)
	assert.Equal(t, "localhost:42069", header["Host"])
	assert.Equal(t, 25, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	header = NewHeader()
	data = []byte("Host:   localhost:42069\r\nFoo: bar\r\n\r\n")
	n, done, err = header.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", header["Host"])
	assert.Equal(t, 25, n)
	assert.False(t, done)
	n, done, err = header.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "bar", header["Foo"])
	assert.Equal(t, 10, n)
	assert.False(t, done)

	// Test: 2 headers with same field name
	header = NewHeader()
	data = []byte("Host:   localhost:42069\r\nHost:     localhost:42070\r\n\r\n")
	n, done, err = header.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 25, n)
	assert.False(t, done)
	n, done, err = header.Parse(data[n:])
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid done
	header = NewHeader()
	data = []byte("Host:   localhost:42069   \r\n\r\n")
	n, done, err = header.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 28, n)
	assert.False(t, done)
	n, done, err = header.Parse(data[n:])
	require.NoError(t, err)
	require.NotNil(t, header)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}
