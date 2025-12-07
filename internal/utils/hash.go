package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(password)); err != nil {
		return "", err
	}
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return false
	}
	return hashedPassword == hash
}
