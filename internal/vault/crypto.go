package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	saltSize = 16
	nonceSize = 12
	keySize = 32
)

var ErrMalformedData = errors.New("malformed encrypted data")

func deriveKey(masterPassword string, salt []byte) []byte {
	return argon2.IDKey([]byte(masterPassword), salt, 1, 64*1024, 4, keySize)
}

func Encrypt(masterPassword, secret string) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := deriveKey(masterPassword, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)

	cipherText := aesgcm.Seal(nil, nonce, []byte(secret), nil)

	blob := make([]byte, 0, saltSize+nonceSize+len(cipherText))
	blob = append(blob, salt...)
	blob = append(blob, nonce...)
	blob = append(blob, cipherText...)
	return blob, nil
}

func Decrypt(masterPassword string, blob []byte) (string, error) {
	if len(blob) < saltSize+nonceSize {
		return "", ErrMalformedData
	}

	salt := blob[:saltSize]
	nonce := blob[saltSize : saltSize+nonceSize]
	cipherText := blob[saltSize+nonceSize:]

	key := deriveKey(masterPassword, salt)

	block, err := aes.NewCipher(key) 
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainBytes, err := aesgcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", errors.New("decryption failed: Incorrect password or corrupted data")
	}
	return string(plainBytes), nil
}