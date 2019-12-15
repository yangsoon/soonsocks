package soonsocks

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Run("config file", func(t *testing.T) {
		config = new(Config)
		configPath := "./testdata/config.json"
		data, err := ioutil.ReadFile(configPath)
		require.Nil(t, err)

		sourceConfig := new(Config)
		err = json.Unmarshal(data, sourceConfig)
		require.Nil(t, err)

		config, err = ParseConfig(configPath)
		require.Nil(t, err)

		assert.Equal(t, sourceConfig.LocalAddr, config.LocalAddr)
		assert.Equal(t, sourceConfig.ServerAddr, config.ServerAddr)
		assert.Equal(t, sourceConfig.Password, config.Password)
		assert.Equal(t, sourceConfig.Method, config.Method)
		assert.Equal(t, sourceConfig.Timeout, config.Timeout)
	})
	
	t.Run("default value", func(t *testing.T) {
		config = new(Config)
		config, err := ParseConfig("")
		require.Nil(t, err)

		assert.Equal(t, defaultServerAddr, config.ServerAddr)
		assert.Equal(t, defaultLocalAddr, config.LocalAddr)
		assert.Equal(t, defaultPassword, config.Password)
		assert.Equal(t, defaultMethod, config.Method)
		assert.Equal(t, defaultTimeout, config.Timeout)
	})
}