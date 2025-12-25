package archive

import "io"

type (
	ExtractOption func(*extractOptions)

	extractOptions struct {
		gzipped         bool
		stripComponents int
	}
)

func Extract(r io.Reader, kind Type, dest string, opts ...ExtractOption) error {
	eOpts := &extractOptions{}
	for _, opt := range opts {
		opt(eOpts)
	}

	switch kind { // nolint: exhaustive
	case Tar:
		return untar(r, dest, eOpts)
	case TarGz:
		eOpts.gzipped = true
		return untar(r, dest, eOpts)
	}

	return nil
}

func StripComponents(n int) ExtractOption {
	return func(e *extractOptions) { e.stripComponents = max(0, n) }
}
