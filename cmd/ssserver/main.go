package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	ss "github.com/yangsoon/soonsocks"
	"io"
	"math/rand"
	"net"
	"strconv"
	"syscall"
	"time"
)

var config *ss.Config

const (
	idxAddrType = 0
	idxIP0      = 1
	idxDmLen    = 1
	idxDm0      = 2

	typeIPv4 = 1
	typeIpv6 = 4
	typeDm   = 3

	lenIPv4   = 1 + net.IPv4len + 2
	lenIPv6   = 1 + net.IPv6len + 2
	lenDmBase = 1 + 1 + 2

	lenHmacSha1 = 10
)

func init() {
	rand.Seed(time.Now().Unix())
}

func getAddressInfo(conn *ss.CConn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))

	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 255(max length address) + 2(port) + 10(hmac-sha1)
	buf := make([]byte, 269)

	if _, err := io.ReadFull(conn, buf[:idxAddrType+1]); err != nil {
		ss.WriteRandomData(conn)
		return "", err
	}

	var reqStart, reqEnd int
	addrType := buf[idxAddrType]
	switch addrType {
	case typeIPv4:
		reqStart, reqEnd = idxIP0, lenIPv4
	case typeIpv6:
		reqStart, reqEnd = idxIP0, lenIPv6
	case typeDm:
		if _, err := io.ReadFull(conn, buf[idxDmLen:idxDmLen+1]); err != nil {
			ss.WriteRandomData(conn)
			return "", nil
		}
		reqStart, reqEnd = idxDm0, lenDmBase+int(buf[idxDmLen])
	default:
		ss.WriteRandomData(conn)
		return "", fmt.Errorf("add type not supported")
	}
	if _, err := io.ReadFull(conn, buf[reqStart: reqEnd]); err != nil {
		ss.WriteRandomData(conn)
		return "", err
	}

	var host string
	switch addrType {
	case typeIPv4:
		host = net.IP(buf[idxIP0: idxIP0+net.IPv4len]).String()
	case typeIpv6:
		host = net.IP(buf[idxIP0: idxIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idxDm0: idxDm0+int(buf[idxDmLen])])
	}

	port := binary.BigEndian.Uint16(buf[reqEnd-2: reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return host, nil
}

func handleConnection(conn *ss.CConn) {
	host, err := getAddressInfo(conn)
	if err != nil {
		ss.Logger.Printf("get host error: %v\n", err)
		return
	}

	ss.Logger.Printf("connecting to host:%v\n", host)

	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			ss.Logger.Printf("dial error: %v\n", err)
		} else {
			ss.Logger.Printf("connecting to %v error: %v\n", host, err)
		}
		return
	}

	go func() {
		defer conn.Close()
		_, err := ss.CopyBuffer(conn, remote)
		if err != nil {
			ss.Logger.Printf("connecting to %v error: %v\n", host, err)
		}
	}()

	_, err = ss.CopyBuffer(remote, conn)
	if err != nil {
		ss.Logger.Printf("connecting to %v error: %v", host, err)
	}
	remote.Close()
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.json", "json file with config")
	flag.Parse()

	var err error
	config, err = ss.ParseConfig(configPath)
	if err != nil {
		ss.Logger.Fatalf("parse %s failed %v \n", configPath, err)
	}
	ss.Logger.Printf("SSServer is running at %v\n", config.ServerAddr)
	ss.Logger.Printf("config info: \n"+
		"--------------------------------\n"+
		"LocalAddr: %v\n"+
		"ServerAddr: %v\n"+
		"Method: %v\n"+
		"--------------------------------\n",
		config.LocalAddr,
		config.ServerAddr,
		config.Method)

	l, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		ss.Logger.Printf("SSServer listen faild %v\n", err)
	}

	cipher, err := ss.NewCipher(config.Method, config.Password)
	if err != nil {
		ss.Logger.Printf("SSServer init cipher error: %v\n", err)
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			ss.Logger.Printf("SSServer accept error %v\n", err)
			continue
		}

		go handleConnection(ss.NewConn(conn, cipher.Clone()))
	}
}
