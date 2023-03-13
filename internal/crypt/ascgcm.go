package crypt

import (
	"crypto/aes"
	"crypto/cipher"
)

func Encrypt(plaintext, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic("[aes new cipher] " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("[cipher new gcm] " + err.Error())
	}

	nonce := make([]byte, 12)
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext
}

func Decrypt(ciphertext, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic("[aes new cipher] " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("[cipher new gcm] " + err.Error())
	}

	nonce := make([]byte, 12)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic("[aesgcm open] " + err.Error())
	}

	return plaintext
}
