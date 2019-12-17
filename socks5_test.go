package soonsocks

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/proxy"
	"net"
	"testing"
)

func TestHandleShake(t *testing.T) {
	config, err := ParseConfig("./testdata/config.json")
	require.Nil(t, err)

	t.Run("test domain address", func(t *testing.T) {
		go func() {
			forward := proxy.FromEnvironment()
			dailer, err := proxy.SOCKS5("tcp", config.LocalAddr, nil, forward)
			require.Nil(t, err)

			conn, err := dailer.Dial("tcp", "yangsoon.github.io:80")
			require.Nil(t, err)
			defer conn.Close()
		}()

		lis, err := net.Listen("tcp", config.LocalAddr)
		require.Nil(t, err)

		for {
			conn, err := lis.Accept()
			assert.Nil(t, err)

			_, host, err := HandleShake(conn)
			require.Nil(t, err)
			assert.Equal(t, "yangsoon.github.io:80", host)
			return
		}
	})

	// TODO test ipv4 ipv6
	//t.Run("test ipv4 address", func(t *testing.T) {
	//	go func() {
	//		forward := proxy.FromEnvironment()
	//		dailer, err := proxy.SOCKS5("tcp", config.LocalAddr, nil, forward)
	//		require.Nil(t, err)
	//
	//		conn, err := dailer.Dial("tcp", "127.0.0.1:80")
	//		require.Nil(t, err)
	//		defer conn.Close()
	//	}()
	//
	//	lis, err := net.Listen("tcp", config.LocalAddr)
	//	require.Nil(t, err)
	//
	//	for {
	//		conn, err := lis.Accept()
	//		require.Nil(t, err)
	//
	//		_, host, err := HandleShake(conn)
	//		require.Nil(t, err)
	//		return
	//	}
	//})
}
