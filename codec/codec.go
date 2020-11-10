package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func Encrypt(key, text []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encrypt: could not encrypt: %w", err)
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(key []byte, text string) ([]byte, error) {
	cipherText, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return nil, fmt.Errorf("decode: error decoding into base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decode: could not create cipher: %w", err)
	}
	if len(text) < aes.BlockSize {
		return nil, fmt.Errorf("decrypt: ciphertext too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	data, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, fmt.Errorf("decode: error decoding string: %w", err)
	}
	return data, nil
}