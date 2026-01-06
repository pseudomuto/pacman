package types

import (
	"database/sql/driver"
	"fmt"
)

const (
	TextFile AssetType = iota
	Archive  AssetType = iota
)

type AssetType int8 // nolint: recvcheck

func (a AssetType) String() string {
	switch a {
	case TextFile:
		return "text"
	case Archive:
		return "archive"
	}

	panic(fmt.Sprintf("unknown asset type: %#v", a))
}

func (a AssetType) Values() []string {
	return []string{
		TextFile.String(),
		Archive.String(),
	}
}

func (a AssetType) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *AssetType) Scan(val any) error {
	var s string
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		s = v
	case []uint8:
		s = string(v)
	}

	switch s {
	case "text":
		*a = TextFile
	case "archive":
		*a = Archive
	}

	if a == nil {
		return fmt.Errorf("unknown asset type: %q", s)
	}

	return nil
}
