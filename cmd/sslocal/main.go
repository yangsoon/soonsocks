package main

import (
	"flag"
	ss "github.com/yangsoon/soonsocks"
	"net"
)

func handleConnection(conn net.Conn) {
	ss.HandleShake(conn)
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.json", "json file with config")
	flag.Parse()

	config, err := ss.ParseConfig(configPath)
	if err != nil {
		ss.Logger.Fatalf("parse %s failed %v \n", configPath, err)
	}
	ss.Logger.Printf("config info: \n" +
		"--------------------------------\n" +
		"LocalAddr: %v\n" +
		"ServerAddr: %v\n" +
		"Method: %v\n" +
		"--------------------------------\n",
		config.LocalAddr,
		config.ServerAddr,
		config.Method)

	l, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		ss.Logger.Printf("SSLocal listen faild %v\n", err)
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			ss.Logger.Printf("SSLocal accept client error: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}