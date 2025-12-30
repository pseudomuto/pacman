package crypto

import (
	"fmt"

	"github.com/google/tink/go/tink"
)

var aeadCipher tink.AEAD

func Encrypt(pt string) ([]byte, error) {
	res, err := aeadCipher.Encrypt([]byte(pt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return res, nil
}

func Decrypt(ct string) ([]byte, error) {
	res, err := aeadCipher.Decrypt([]byte(ct), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return res, nil
}

func SetCipher(c tink.AEAD) {
	aeadCipher = c
}
