package crypt

import (
	"crypto/md5"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

const (
	pwdCost = 5
)

// GetMD5hash is a helper function for generate MD5 hash.
func GetMD5hash(data string) []byte {
	h := md5.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

// GetMD5hash is a helper function for generate SHA256 hash.
func GetSHA256hash(data string) []byte {
	h := sha256.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

// CalculatePasswordHash is a helper function for generate hash from password.
func CalculatePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), pwdCost)
	return string(bytes), err
}

// CheckPasswordHash is a helper function used to compare password and hash.
// Returns nil on success, or an error on failure.
func CheckPasswordHashStr(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
