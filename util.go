package soonsocks

import (
	"crypto/md5"
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "SSLogger: ", log.Ldate | log.Ltime)

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