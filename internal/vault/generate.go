package vault

import (
	"crypto/rand"
	"math/big"
)

const (
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	numberChars = "0123456789"
	symbolChars = "!@#$%^&*()-_=+,.?/:;{}[]~"
)

func GeneratePassword(length int, useSymbols bool) (string, error) {
	charset := upperChars + lowerChars + numberChars
	if useSymbols {
		charset += symbolChars
	}

	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomIndex.Int64()]
	}
	return string(password), nil
}