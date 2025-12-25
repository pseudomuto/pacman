package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func tarDir(w io.Writer, src string, opts *compressOptions) error {
	out := w
	if opts.gzipped {
		out = gzip.NewWriter(w)
		defer func() { _ = out.(io.WriteCloser).Close() }()
	}

	tw := tar.NewWriter(out)
	defer func() { _ = tw.Close() }()

	if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			header.Name = filepath.Base(src)
		} else {
			header.Name = relPath
		}

		if d.IsDir() {
			header.Name += "/"
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !d.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to compress dir: %s, %w", src, err)
	}

	return nil
}

func untar(r io.Reader, dest string, opts *extractOptions) error {
	in := io.NopCloser(r)
	if opts.gzipped {
		var err error
		if in, err = gzip.NewReader(r); err != nil {
			return fmt.Errorf("failed to open gzip stream: %w", err)
		}
	}

	defer func() { _ = in.Close() }()

	rdr := tar.NewReader(in)
	for {
		hdr, err := rdr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		path := hdr.Name
		for range opts.stripComponents {
			path = path[strings.Index(path, "/")+1:]
		}

		path = filepath.Join(dest, path) //nolint:gosec // we handle ZipSlip below
		if strings.HasPrefix(path, filepath.Clean(path)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path: %s", path)
		}

		if hdr.Typeflag == tar.TypeReg {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %s, %w", dir, err)
			}

			out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create file: %s, %w", path, err)
			}

			if _, err := io.Copy(out, rdr); err != nil { //nolint:gosec
				return fmt.Errorf("failed to write file: %s, %w", path, err)
			}

			if err := out.Close(); err != nil {
				return fmt.Errorf("failed to close file: %s, %w", path, err)
			}
		}
	}

	return nil
}
