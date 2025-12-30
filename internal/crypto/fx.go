package crypto

import (
	"fmt"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"go.uber.org/fx"
)

var Module = fx.Module("crypto",
	fx.Provide(ReadKey),
	fx.Invoke(func(ksh *keyset.Handle) error {
		cipher, err := aead.New(ksh)
		if err != nil {
			return fmt.Errorf("failed to create cipher: %w", err)
		}

		aeadCipher = cipher
		return nil
	}),
)
