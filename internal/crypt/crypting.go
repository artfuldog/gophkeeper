// Package crypting contains functions for encrypting and decrypting messages.
package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	// 32 bytes for AES-256.
	AESKeyLength = 32
)

// EncryptAES encrypts message with AES GCM.
// The key argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func EncryptAES(key, message []byte) ([]byte, error) {
	gcm, err := getGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, message, nil), nil
}

// EncryptAES encrypts message with AES GCM and additional information.
// The key argument can be any length, for key streching to 16, 24, or 32 bytes
// will be used derivation function (deriveKey).
func EncryptAESwithAD(key, message []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, err
	}

	gcm, err := getGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, message, nil)
	encrypted = append(encrypted, salt...)

	return encrypted, nil
}

// DecryptAES decrypts encrypted with AES GCM message.
// The key argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func DecryptAES(key, encrypted []byte) ([]byte, error) {
	gcm, err := getGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("to short encrypted message")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// DecryptAESwithAD decrypts encrypted with AES GCM message and additional information.
// The key argument can be any length, for key stretching to 16, 24, or 32 bytes
// will be used derivation function (deriveKey).
func DecryptAESwithAD(key, encrypted []byte) ([]byte, error) {
	if len(encrypted) < AESKeyLength {
		return nil, errors.New("to short encrypted message")
	}

	salt, encrypted := encrypted[len(encrypted)-AESKeyLength:], encrypted[:len(encrypted)-AESKeyLength]

	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, err
	}

	gcm, err := getGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// deriveKey is a key derivation function which stretch the password to make it
// suitable for user as cryptographic key.
func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, AESKeyLength)
		rand.Read(salt) //nolint:errcheck
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, AESKeyLength)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

// getGCM is a helper function which returns block cipher wrapped in Galois Counter Mode
// with the standard nonce length.
func getGCM(key []byte) (cipher.AEAD, error) {
	cb, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cb)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

// GenerateRandomKey32 generates random 32-bytes length key.
func GenerateRandomKey32() []byte {
	key := make([]byte, AESKeyLength)
	rand.Read(key) //nolint:errcheck

	return key
}
