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