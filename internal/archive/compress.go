package archive

import "io"

type (
	CompressOption func(*compressOptions)

	compressOptions struct {
		gzipped          bool
		prefixComponents []string
	}
)

func Compress(w io.Writer, kind Type, src string, opts ...CompressOption) error {
	copts := new(compressOptions)
	for _, opt := range opts {
		opt(copts)
	}

	switch kind { // nolint: exhaustive
	case Tar:
		return tarDir(w, src, copts)
	case TarGz:
		copts.gzipped = true
		return tarDir(w, src, copts)
	}

	return nil
}

func PrefixComponents(dirs ...string) CompressOption {
	return func(co *compressOptions) { co.prefixComponents = dirs }
}
