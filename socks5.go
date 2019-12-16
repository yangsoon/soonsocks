package soonsocks

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	socksVersion    = 5
	socksCmdConnect = 1
)

var (
	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
)

type (
	Socks5Negotiation struct {
		Version      uint8
		NumOfMethods uint8
		Methods      []uint8
	}

	Socks5Request struct {
		Version     uint8
		Command     uint8
		RSV         uint8
		AddressType uint8
		Address     string
		Port        uint16
		Host        string
		RawAddr     []byte
	}
)

func extractNegotiation(conn net.Conn) (socks5n Socks5Negotiation, err error) {

	const (
		idxVer        = 0
		idxNofMethods = 1
		idxMethods    = 2
	)

	buf := make([]byte, 258)

	var n int
	// make sure the nmethod field
	if n, err = io.ReadAtLeast(conn, buf, idxNofMethods+1); err != nil {
		return
	}

	if buf[idxVer] != socksVersion {
		err = errVer
		return
	}

	nmethods := int(buf[idxNofMethods])
	msgLen := nmethods + 2
	if n == msgLen {
		// do nothing
	} else if n < msgLen {
		if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
			return
		}
	} else {
		err = errAuthExtraData
		return
	}

	// only support one method
	socks5n = Socks5Negotiation{
		Version: socksVersion,
		NumOfMethods: idxNofMethods,
		Methods: []byte{buf[idxMethods]},
	}

	return
}

func replyNegotiation(conn net.Conn) (err error) {
	_, err = conn.Write([]byte{socksVersion, 0x00})
	return
}

func extractRequest(conn net.Conn) (socks5r Socks5Request, err error) {

	const (
		idxVer      = 0
		idxCmd      = 1
		idxRSV      = 2
		idxAddrType = 3
		idxIP0      = 4
		idxDmLen    = 4
		idxDm0      = 5

		typeIPv4 = 1
		typeIPv6 = 4
		typeDm   = 3

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)

	buf := make([]byte, 263)
	var n int

	conn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
	if n, err = io.ReadAtLeast(conn, buf, idxDmLen+1); err != nil {
		return
	}

	if buf[idxVer] != socksVersion {
		err = errVer
		return
	}

	if buf[idxCmd] != socksCmdConnect {
		err = errCmd
		return
	}

	reqLen := -1
	switch buf[idxAddrType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idxDmLen]) + lenDmBase
	}

	if n == reqLen {
		// common case do noting
	} else if n < reqLen {
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
	} else {
		err = errReqExtraData
		return
	}

	var address string
	rawaddr := buf[idxAddrType:reqLen]

	switch buf[idxAddrType] {
	case typeIPv4:
		address = net.IP(buf[idxIP0:idxIP0+lenIPv4]).String()
	case typeIPv6:
		address = net.IP(buf[idxIP0:idxIP0+lenIPv6]).String()
	case typeDm:
		address = string(buf[idxDm0 : idxDm0+buf[idxDmLen]])
	}

	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])

	socks5r = Socks5Request{
		Version:     socksVersion,
		Command:     buf[idxCmd],
		RSV:         buf[idxRSV],
		AddressType: buf[idxAddrType],
		Address:     address,
		Port:        port,
		Host:        net.JoinHostPort(address, strconv.Itoa(int(port))),
		RawAddr:     rawaddr,
	}
	return
}

func replyRequest(conn net.Conn) (err error) {
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x08, 0x43})
	return
}

func HandleShake(conn net.Conn) (rawaddr []byte, host string, err error) {

	rawaddr = []byte{}
	host = ""

	// 1. get pkg from client
	if _, err = extractNegotiation(conn); err != nil {
		return
	}
	Logger.Println("get conn from client")

	// 2. reply to client build connect
	if err = replyNegotiation(conn); err != nil {
		return
	}
	Logger.Println("reply to client")

	// 3. get request pkg from client
	var socks5r Socks5Request
	if socks5r, err = extractRequest(conn); err != nil {
		return
	}
	Logger.Printf("request %s\n", socks5r.Host)

	// 4. reply to client
	if err = replyRequest(conn); err != nil {
		return
	}
	Logger.Println("reply to client request")

	rawaddr = socks5r.RawAddr
	host = socks5r.Host
	return
}

