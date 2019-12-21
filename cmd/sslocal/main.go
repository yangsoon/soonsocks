package main

import (
	"flag"
	ss "github.com/yangsoon/soonsocks"
	"math/rand"
	"net"
	"time"
)

var config *ss.Config


func init() {
	rand.Seed(time.Now().Unix())
}


func handleConnection(conn net.Conn) {
	rawaddr, host, err := ss.HandleShake(conn)
	if err != nil {
		ss.Logger.Printf("socks negotiate host %s error: %v\n", host, err)
		return
	}

	cipher, err := ss.NewCipher(config.Method, config.Password)
	if err != nil {
		ss.Logger.Printf("create cipher error: %v\n", err)
		return
	}

	serverCConn, err := ss.DialWithCipher(config.ServerAddr, cipher.Clone())
	if err != nil {
		ss.Logger.Printf("connect to server %s error: %v\n", config.ServerAddr, err)
		return
	}
	defer serverCConn.Close()

	_, err = serverCConn.Write(rawaddr)
	if err != nil {
		ss.Logger.Printf("write to server %s error: %v\n", config.ServerAddr, err)
	}

	go func() {
		defer conn.Close()
		ss.CopyBuffer(conn, serverCConn)
	}()

	ss.CopyBuffer(serverCConn, conn)
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