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

		var tests = []struct{
			expected interface{}
			actual interface{}
		}{
			{sourceConfig.LocalAddr, config.LocalAddr},
			{sourceConfig.ServerAddr, config.ServerAddr},
			{sourceConfig.Password, config.Password},
			{sourceConfig.Method, config.Method},
			{sourceConfig.Timeout, config.Timeout},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, test.actual)
		}
	})
	
	t.Run("default value", func(t *testing.T) {
		config = new(Config)
		config, err := ParseConfig("")
		require.Nil(t, err)

		var tests = []struct{
			expected interface{}
			actual interface{}
		}{
			{defaultLocalAddr, config.LocalAddr},
			{defaultServerAddr, config.ServerAddr},
			{defaultPassword, config.Password},
			{defaultMethod, config.Method},
			{defaultTimeout, config.Timeout},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, test.actual)
		}
	})
}