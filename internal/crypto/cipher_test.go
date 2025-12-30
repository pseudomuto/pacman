package crypto_test

import (
	"testing"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	. "github.com/pseudomuto/pacman/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)

	cipher, err := aead.New(kh)
	require.NoError(t, err)
	SetCipher(cipher) // NB: can't use t.Parallel

	secret := "ssshhhhh"

	ct, err := Encrypt(secret)
	require.NoError(t, err)

	pt, err := Decrypt(string(ct))
	require.NoError(t, err)

	require.Equal(t, secret, string(pt))
}
