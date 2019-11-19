package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

const (
	PasswordLength    int   = 32
)

func hashPassphraseToFixLength(input []byte) []byte {
	sha_256 := sha256.New()
	sha_256.Write(input)
	result := sha_256.Sum(nil)
	return result[:PasswordLength]
}

func EncryptData(data, passphrase []byte) ([]byte, []byte, error) {
	key := hashPassphraseToFixLength(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	cipherdata := make([]byte, len(data))
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}, []byte{}, err
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherdata, data)
	return cipherdata, iv, nil
}

func DecryptData(cipherdata, passphrase, iv []byte) ([]byte, error) {
	key := hashPassphraseToFixLength(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	data := make([]byte, len(cipherdata))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(data, cipherdata)
	return data, nil
}