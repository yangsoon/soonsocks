package soonsocks

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCipher(t *testing.T) {
	t.Run("should return error", func(t *testing.T) {
		cipher, err := NewCipher("unsupport method", "password")
		assert.Nil(t, cipher)
		assert.NotNil(t, err)
	})
	
	t.Run("cipher aes-128-cfb", func(t *testing.T) {
		cipher, err := NewCipher("aes-128-cfb", "password")
		require.Nil(t, err)
		assert.Equal(t, 16, cipher.info.keyLen)
		assert.Equal(t, 16, cipher.info.ivLen)
		assert.Equal(t, generateKey("password", 16), cipher.key)
		assert.Nil(t, cipher.enc)
		assert.Nil(t, cipher.dec)

		require.Nil(t, cipher.initEncrypt())
		assert.Equal(t, len(cipher.iv), cipher.info.ivLen)
		plaintext := []byte("plain text")
		encrypttext := make([]byte, len(plaintext))
		cipher.Encrypt(encrypttext, plaintext)
		require.Nil(t, cipher.initDecrypt(cipher.iv))
		decrypttext := make([]byte, len(plaintext))
		cipher.Decrypt(decrypttext, encrypttext)
		assert.Equal(t, plaintext, decrypttext)

		newCipher := cipher.Clone()
		assert.Nil(t, newCipher.dec)
		assert.Nil(t, newCipher.enc)
	})

	t.Run("cipher aes-256-cfb", func(t *testing.T) {
		cipher, err := NewCipher("aes-256-cfb", "password")
		require.Nil(t, err)
		assert.Equal(t, 32, cipher.info.keyLen)
		assert.Equal(t, 16, cipher.info.ivLen)
		assert.Equal(t, generateKey("password", 32), cipher.key)
		assert.Nil(t, cipher.enc)
		assert.Nil(t, cipher.dec)

		require.Nil(t, cipher.initEncrypt())
		assert.Equal(t, len(cipher.iv), cipher.info.ivLen)
		plaintext := []byte("plain text")
		encrypttext := make([]byte, len(plaintext))
		cipher.Encrypt(encrypttext, plaintext)
		require.Nil(t, cipher.initDecrypt(cipher.iv))
		decrypttext := make([]byte, len(plaintext))
		cipher.Decrypt(decrypttext, encrypttext)
		assert.Equal(t, plaintext, decrypttext)

		newCipher := cipher.Clone()
		assert.Nil(t, newCipher.dec)
		assert.Nil(t, newCipher.enc)
	})

	t.Run("cipher rc4-md5", func(t *testing.T) {
		cipher, err := NewCipher("rc4-md5", "password")
		assert.Nil(t, err)
		assert.NotNil(t, cipher)
		assert.Equal(t, 16, cipher.info.keyLen)
		assert.Equal(t, 16, cipher.info.ivLen)
		assert.Equal(t, generateKey("password", 16), cipher.key)
		assert.Nil(t, cipher.enc)
		assert.Nil(t, cipher.dec)

		require.Nil(t, cipher.initEncrypt())
		assert.Equal(t, len(cipher.iv), cipher.info.ivLen)
		plaintext := []byte("test data")
		encrypttext := make([]byte, len(plaintext))
		cipher.Encrypt(encrypttext, plaintext)
		require.Nil(t, cipher.initDecrypt(cipher.iv))
		decrypttext := make([]byte, len(plaintext))
		cipher.Decrypt(decrypttext, encrypttext)
		assert.Equal(t, plaintext, decrypttext)

		newCipher := cipher.Clone()
		assert.Nil(t, newCipher.dec)
		assert.Nil(t, newCipher.enc)
	})
}