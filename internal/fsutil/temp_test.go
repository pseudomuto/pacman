package fsutil_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/pseudomuto/pacman/internal/fsutil"
	"github.com/stretchr/testify/require"
)

func TestWithTempDir(t *testing.T) {
	t.Run("creates writable directory that gets cleaned up", func(t *testing.T) {
		var capturedPath string

		require.NoError(t, WithTempDir(func(dir string) error {
			capturedPath = dir

			// Verify directory exists
			info, err := os.Stat(dir)
			require.NoError(t, err)
			require.True(t, info.IsDir())

			// Verify directory is writable
			testFile := filepath.Join(dir, "test.txt")
			require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

			return nil
		}))

		// Verify directory was cleaned up
		_, err := os.Stat(capturedPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("propagates callback error", func(t *testing.T) {
		expectedErr := errors.New("callback error")

		require.ErrorIs(t, WithTempDir(func(dir string) error {
			return expectedErr
		}), expectedErr)
	})

	t.Run("cleans up directory even when callback returns error", func(t *testing.T) {
		var capturedPath string

		_ = WithTempDir(func(dir string) error {
			capturedPath = dir
			return errors.New("some error")
		})

		// Verify directory was cleaned up despite error
		_, err := os.Stat(capturedPath)
		require.True(t, os.IsNotExist(err))
	})
}

func TestWithTempFile(t *testing.T) {
	t.Run("creates writable file that gets cleaned up", func(t *testing.T) {
		var capturedPath string

		require.NoError(t, WithTempFile(func(f *os.File) error {
			capturedPath = f.Name()

			// Verify file exists and is writable
			_, err := f.WriteString("test content")
			require.NoError(t, err)

			return nil
		}))

		// Verify file was cleaned up
		_, err := os.Stat(capturedPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("propagates callback error", func(t *testing.T) {
		expectedErr := errors.New("callback error")

		require.ErrorIs(t, WithTempFile(func(f *os.File) error {
			return expectedErr
		}), expectedErr)
	})

	t.Run("provides valid open file handle", func(t *testing.T) {
		require.NoError(t, WithTempFile(func(f *os.File) error {
			// File should be open and valid
			require.NotNil(t, f)

			// Should be able to get file info
			info, err := f.Stat()
			require.NoError(t, err)
			require.False(t, info.IsDir())

			return nil
		}))
	})
}
