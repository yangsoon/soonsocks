package soonsocks

import (
	"errors"
	"io"
	"net"
)

// Conn With Cipher
type CConn struct {
	net.Conn
	*Cipher
}

func NewConn(conn net.Conn, cipher *Cipher) *CConn {
	return &CConn{
		conn,
		cipher,
	}
}

func (cc *CConn) Read(b []byte) (int, error){
	if cc.dec == nil {
		iv := make([]byte, cc.info.ivLen)
		if _, err := io.ReadFull(cc.Conn, iv); err != nil {
			return 0, err
		}
		if err := cc.initDecrypt(iv); err!=nil {
			return 0, err
		}

		if len(cc.iv) == 0 {
			cc.iv = iv
		}
	}
	encryptData := make([]byte, len(b))
	n, err := cc.Conn.Read(encryptData)
	if n > 0 {
		cc.Decrypt(b[0:n], encryptData[0:n])
	}
	return n, err
}

func (cc *CConn) Write(b []byte) (int, error) {
	var iv []byte
	if cc.enc == nil {
		if err := cc.initEncrypt(); err != nil {
			return 0, err
		}
		if len(cc.iv) == 0 {
			return 0, errors.New("get iv error")
		}
		iv = cc.iv
	}

	encryptData := make([]byte, len(iv)+len(b))
	if len(iv) > 0 {
		copy(encryptData, iv)
	}

	cc.Encrypt(encryptData[len(iv):], b)
	n, err := cc.Conn.Write(encryptData)
	return n, err
}