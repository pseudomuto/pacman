package publisher_test

//go:generate go tool mockgen -destination=mocks_test.go -package=publisher_test . Packager,Uploader,VCSFetcher

import (
	"io"
	"testing"

	"github.com/pseudomuto/pacman/internal/archive"
	"github.com/pseudomuto/pacman/internal/packager"
	. "github.com/pseudomuto/pacman/internal/publisher"
	"github.com/pseudomuto/pacman/internal/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPublisher_Publish(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := NewMockVCSFetcher(ctrl)
	uploader := NewMockUploader(ctrl)

	t.Run("go module", func(t *testing.T) {
		publisher := New(PublisherParams{
			Packagers:   []Packager{packager.NewGoModule()},
			Uploaders:   []Uploader{uploader},
			VCSFetchers: []VCSFetcher{fetcher},
		})

		pubOpts := PublishOptions{
			Type:    types.GoModule,
			Storage: types.GCS,
			VCS:     types.GitHub,
			Repo:    "test/repo",
			Ref:     "abcdef12345",
			Subdir:  "sub/dir/project",
			Package: "github.com/pseudomuto/test",
			Version: "v1.2.3",
		}

		fetcher.EXPECT().Type().Return(pubOpts.VCS)
		uploader.EXPECT().Type().Return(pubOpts.Storage)

		fetcher.EXPECT().
			FetchArchive(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(w io.Writer, repo string, opts types.VCSOptions) error {
				require.Equal(t, pubOpts.Repo, repo)
				require.Equal(t, types.VCSOptions{
					Ref: pubOpts.Ref,
					Dir: pubOpts.Subdir,
				}, opts)

				return archive.Compress(
					w,
					archive.TarGz,
					"../../testdata/gomodule",
					archive.PrefixComponents(
						"repo",
						"sub",
						"dir",
						"project",
					),
				)
			})

		uploader.EXPECT().Write(
			gomock.Any(),
			gomock.Any(),
			"gomod/github.com/pseudomuto/test@v1.2.3.zip",
		).Return(nil)

		require.NoError(t, publisher.Publish(t.Context(), pubOpts))
	})

	t.Run("misconfigured", func(t *testing.T) {
		packager := NewMockPackager(ctrl)
		publisher := New(PublisherParams{
			Packagers:   []Packager{packager},
			Uploaders:   []Uploader{uploader},
			VCSFetchers: []VCSFetcher{fetcher},
		})

		t.Run("unknown packager", func(t *testing.T) {
			packager.EXPECT().Type().Return(types.GoModule)
			require.EqualError(t, publisher.Publish(t.Context(), PublishOptions{
				Type: types.ArchiveType(100),
			}), "unknown packager: 100")
		})

		t.Run("unknown VCS fetcher", func(t *testing.T) {
			packager.EXPECT().Type().Return(types.GoModule)
			fetcher.EXPECT().Type().Return(types.GitHub)

			require.EqualError(t, publisher.Publish(t.Context(), PublishOptions{
				Type: types.GoModule,
				VCS:  types.VCSType(100),
			}), "unknown fetcher: 100")
		})

		t.Run("unknown uploader", func(t *testing.T) {
			packager.EXPECT().Type().Return(types.GoModule)
			fetcher.EXPECT().Type().Return(types.GitHub)
			uploader.EXPECT().Type().Return(types.GCS)

			require.EqualError(t, publisher.Publish(t.Context(), PublishOptions{
				Type:    types.GoModule,
				VCS:     types.GitHub,
				Storage: types.StorageType(100),
			}), "unknown uploader: 100")
		})
	})
}
