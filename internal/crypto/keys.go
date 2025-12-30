package crypto

import (
	"fmt"
	"os"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"github.com/pseudomuto/pacman/internal/config"
)

// masterKey is a dummy key used as Tink's root key. In real life, this would be a KMS key or something better than an
// empty value.
//
// TODO: Make this better
type masterKey struct{}

// CreateKey creates a new AEAD (256-bit AES) key and writes it to the specified path.
func CreateKey(path string) (*keyset.Handle, error) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		return nil, fmt.Errorf("failed to create handle: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %s, %w", path, err)
	}
	defer func() { _ = file.Close() }()

	bin := keyset.NewBinaryWriter(file)
	if err := kh.Write(bin, new(masterKey)); err != nil {
		return nil, fmt.Errorf("failed to write keysey: %w", err)
	}

	return kh, nil
}

// ReadKey reads the CryptoKey specified by the supplied Config. This key will be used for encrypting/decrypting
// database secrets.
func ReadKey(c *config.Config) (*keyset.Handle, error) {
	file, err := os.Open(c.CryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to open cryptoKey: %s, %w", c.CryptoKey, err)
	}

	bin := keyset.NewBinaryReader(file)
	kh, err := keyset.Read(bin, new(masterKey))
	if err != nil {
		return nil, fmt.Errorf("failed to read cryptoKey: %w", err)
	}

	return kh, nil
}

func (k *masterKey) Encrypt(pt, data []byte) ([]byte, error) {
	return pt, nil
}

func (k *masterKey) Decrypt(ct, data []byte) ([]byte, error) {
	return ct, nil
}
