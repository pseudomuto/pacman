package crypto_test

import (
	"path/filepath"
	"testing"

	"github.com/google/tink/go/aead"
	"github.com/pseudomuto/pacman/internal/config"
	. "github.com/pseudomuto/pacman/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestReadWriteKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "keys.bin")
	secret := []byte("ssshhh")

	wkh, err := CreateKey(file)
	require.NoError(t, err)

	cipher, err := aead.New(wkh)
	require.NoError(t, err)

	ct, err := cipher.Encrypt(secret, nil)
	require.NoError(t, err)

	rkh, err := ReadKey(&config.Config{
		CryptoKey: file,
	})
	require.NoError(t, err)

	cipher, err = aead.New(rkh)
	require.NoError(t, err)

	pt, err := cipher.Decrypt(ct, nil)
	require.NoError(t, err)

	require.Equal(t, secret, pt)
}
