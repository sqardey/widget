package http_transport

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
)

func (h *Handler) InitEncryption() error {
	key, err := base64.StdEncoding.DecodeString(
		os.Getenv("CREDENTIALS_KEY"),
	)
	if err != nil {
		return fmt.Errorf("неверный формат CREDENTIALS_KEY: %w", err)
	}
	if len(key) != 32 {
		return fmt.Errorf("CREDENTIALS_KEY должен содержать 32 байта в Base64")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	h.credentialsCipher, err = cipher.NewGCMWithRandomNonce(block)
	return err
}

func (h *Handler) encrypt(value string) string {
	encrypted := h.credentialsCipher.Seal(
		nil, nil, []byte(value), nil,
	)
	return base64.StdEncoding.EncodeToString(encrypted)
}

func (h *Handler) decrypt(value string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	decrypted, err := h.credentialsCipher.Open(
		nil, nil, encrypted, nil,
	)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
