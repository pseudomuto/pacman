package crypto

import (
	"database/sql/driver"
	"encoding/base64"
	"fmt"
)

// Secret defines a database secret.
//
// It will be encrypted when written, and decrypted when read.
type Secret string //nolint: recvcheck

// Scan reads the value from the db.
func (s *Secret) Scan(v any) error {
	if v == nil {
		return nil
	}

	var ct string
	switch val := v.(type) {
	case string:
		ct = val
	case []byte:
		ct = string(val)
	default:
		return fmt.Errorf("invalid type for secret: %T", val)
	}

	val, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		return fmt.Errorf("failed to decode secret: %w", err)
	}

	pt, err := Decrypt(string(val))
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	*s = Secret(pt)
	return nil
}

// Value defines the value to be inserted into the db.
func (s Secret) Value() (driver.Value, error) {
	bytes, err := Encrypt(string(s))
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}
