package soonsocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rc4"
)

type Cipher struct {
	enc cipher.Stream
	dec cipher.Stream
	key []byte
	iv []byte
	info *CipherInfo
	method string
}

type CipherInfo struct {
	keyLen int
	ivLen int
	newStream func([]byte, []byte, bool)(cipher.Stream, error)
}

var cipherMethods = map[string]*CipherInfo{
	"aes-128-cfb": &CipherInfo{16, 16, newAESCFBStream},
	"aes-256-cfb": &CipherInfo{32, 16, newAESCFBStream},
	"rc4-md5":     &CipherInfo{16, 16, newRC4MD5Stream},
}

func newRC4MD5Stream(key, iv []byte, _ bool) (cipher.Stream, error) {
	rc4key := md5sum(append(key, iv...))
	return rc4.NewCipher(rc4key)
}

func newAESCFBStream(key, iv []byte, isEncrypt bool) (cipher.Stream, error){
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if isEncrypt {
		return cipher.NewCFBEncrypter(block, iv), nil
	}
	return cipher.NewCFBDecrypter(block, iv), nil
}

func NewCipher(method, passwd string) (*Cipher, error) {

}