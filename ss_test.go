package soonsocks

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSSLocal(t *testing.T) {
	configPath := "./testdata/config.json"
	config, err := ParseConfig(configPath)

	require.Nil(t, err)

	dialer, err := proxy.SOCKS5("tcp", config.LocalAddr, nil, &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	})

	transport := &http.Transport{
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
	}

	response, err := client.Get("http://www.google.com/")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	assert.Equal(t, "200 OK", response.Status, "The SSLocal couldn't use")
}
