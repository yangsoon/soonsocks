package soonsocks

import (
	"encoding/json"
	"io/ioutil"
)

var (
	defaultServerAddr = "0.0.0.0:7892"
	defaultLocalAddr = "0.0.0.0:7891"
	defaultPassword = "password"
	defaultMethod = "aes-128-cfb"
	defaultTimeout = 300

	config  = new(Config)
)

type Config struct {
	LocalAddr string `json:"local_addr"`
	ServerAddr string `json:"server_addr"`
	Password string `json:"password"`
	Method string `json:"method"`
	Timeout int `json:"timeout"`
}

func ParseConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	if config.ServerAddr == "" {
		config.ServerAddr = defaultServerAddr
	}

	if config.LocalAddr == "" {
		config.LocalAddr = defaultLocalAddr
	}

	if config.Password == "" {
		config.Password = defaultPassword
	}

	if config.Method == "" {
		config.Method = defaultMethod
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return config, nil
}