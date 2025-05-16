package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"quokka-ai-bot/config"
)

var encryption_key = []byte(config.Load().AesKey) // openssl rand -hex 32

func EncryptMessage(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryption_key) // creating a cipher
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block) // set up a mode that guarantees data integrity
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptMessage(encryptedtext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedtext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryption_key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
