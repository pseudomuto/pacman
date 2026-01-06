package goproxy_test

import (
	"os"
	"testing"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"github.com/pseudomuto/pacman/internal/crypto"
)

func TestMain(m *testing.M) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		panic(err)
	}

	cipher, err := aead.New(kh)
	if err != nil {
		panic(err)
	}

	crypto.SetCipher(cipher)

	os.Exit(m.Run())
}
