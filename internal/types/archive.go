package types

import (
	"database/sql/driver"
	"fmt"
)

const (
	GoModule ArchiveType = iota
)

type ArchiveType int8 // nolint: recvcheck

func (a ArchiveType) String() string {
	switch a {
	case GoModule:
		return "gomod"
	}

	panic(fmt.Sprintf("unknown archive type: %d", a))
}

func (a ArchiveType) Values() []string {
	return []string{
		GoModule.String(),
	}
}

func (a ArchiveType) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *ArchiveType) Scan(val any) error {
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
	case "gomod":
		*a = GoModule
	}

	if a == nil {
		return fmt.Errorf("unknown archive type: %q", s)
	}

	return nil
}
