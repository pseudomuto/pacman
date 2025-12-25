package archive

import "fmt"

const (
	Tar   Type = iota
	TarGz Type = iota
	Zip   Type = iota
)

type Type uint8

func (a Type) String() string {
	switch a {
	case Tar:
		return "tar"
	case TarGz:
		return "tar.gz"
	case Zip:
		return "zip"
	}

	panic(fmt.Sprintf("unknown archive type: %d", a))
}
