package soonsocks

import "net"

func DialWithCipher(address string, cipher *Cipher) (*CConn, error){
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return NewConn(conn, cipher), nil
}