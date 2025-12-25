package types

import "fmt"

const (
	FileSystem StorageType = iota
	GCS        StorageType = iota
)

type StorageType uint8

func (s StorageType) String() string {
	switch s {
	case FileSystem:
		return "fs"
	case GCS:
		return "gcs"
	}

	panic(fmt.Sprintf("unknown storage type: %T", s))
}
