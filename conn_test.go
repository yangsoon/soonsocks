package soonsocks

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"testing"
)

func TestNewConn(t *testing.T) {
	config, err := ParseConfig("./testdata/config.json")
	require.Nil(t, err)

	message := bytes.Repeat([]byte("test message"), 1<<10)
	cipher, err := NewCipher("aes-128-cfb", "password")

	assert.Nil(t, err)

	go func() {
		conn, err := net.Dial("tcp", config.ServerAddr)

		assert.Nil(t, err)
		serverCConn := NewConn(conn, cipher)
		defer serverCConn.Close()

		_, err = serverCConn.Write(message)
		assert.Nil(t, err)
	}()

	l, err := net.Listen("tcp", config.ServerAddr)
	assert.Nil(t, err)

	for {
		conn, err := l.Accept()
		assert.Nil(t, err)
		clientCCon := NewConn(conn, cipher)
		defer clientCCon.Close()

		buf, err := ioutil.ReadAll(clientCCon)
		assert.Nil(t, err)
		assert.Equal(t, message, buf)
		return
	}
}
