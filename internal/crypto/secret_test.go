package crypto_test

import (
	"testing"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	. "github.com/pseudomuto/pacman/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestSecret_Scan(t *testing.T) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)

	cipher, err := aead.New(kh)
	require.NoError(t, err)
	SetCipher(cipher) // NB: can't use t.Parallel

	secret := Secret("ssshhhhh")

	t.Run("strings", func(t *testing.T) {
		t.Parallel()

		res, err := secret.Value()
		require.NoError(t, err)

		var s Secret
		require.NoError(t, (&s).Scan(res))
		require.Equal(t, secret, s)
	})

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()

		res, err := secret.Value()
		require.NoError(t, err)

		var s Secret
		require.NoError(t, (&s).Scan([]byte(res.(string))))
		require.Equal(t, secret, s)
	})

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()

		var s Secret
		require.NoError(t, (&s).Scan(nil))
	})

	t.Run("invalid type", func(t *testing.T) {
		var s Secret
		require.ErrorContains(t, (&s).Scan(0), "invalid type for secret")
	})

	t.Run("bad value", func(t *testing.T) {
		var s Secret
		require.ErrorContains(t, (&s).Scan("not b64"), "failed to decode secret")
	})
}

func TestSecret_Value(t *testing.T) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)

	cipher, err := aead.New(kh)
	require.NoError(t, err)
	SetCipher(cipher) // NB: can't use t.Parallel

	secret := Secret("ssshhhhh")
	res, err := secret.Value()
	require.NoError(t, err)

	str, ok := res.(string)
	require.True(t, ok)
	require.NotEqual(t, string(secret), str)
}
