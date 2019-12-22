package soonsocks

import (
	"crypto/md5"
	"io"
	"log"
	"os"
	mrand "math/rand"
	"crypto/rand"
)

var Logger = log.New(os.Stdout, "SSLogger: ", log.Ldate | log.Ltime| log.Lshortfile | log.LstdFlags)

func md5sum(b []byte) []byte {
	h := md5.New()
	h.Write(b)
	return h.Sum(nil)
}

func generateKey(password string, keyLen int) []byte {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)

	copy(m, md5sum([]byte(password)))

	// Repeatedly call md5 until bytes generated is enough.
	// Each call to md5 uses data: prev md5 sum + password.
	// The result of md5sum is 16byte(128bit)
	// So need to filling the m to the keyLen length
	d := make([]byte, md5Len + len(password))
	start := 0
	for i:=1; i < cnt; i++ {
		start += md5Len
		copy(d, m[start-md5Len: start])
		copy(d[md5Len:], password)
		copy(m[start:], md5sum(d))
	}
	return m[:keyLen]
}

func CopyBuffer(dst io.Writer, src io.Reader) (n int64, err error) {
	buf := make([]byte, 32*1024)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				n += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func WriteRandomData(conn io.Writer) error {
	rlen := mrand.Int() % 8767
	data := make([]byte, rlen)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}