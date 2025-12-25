package types

import "fmt"

const (
	GoModule ArchiveType = iota
)

type ArchiveType int8

func (a ArchiveType) String() string {
	switch a {
	case GoModule:
		return "gomod"
	}

	panic(fmt.Sprintf("unknown archive type: %T", a))
}
