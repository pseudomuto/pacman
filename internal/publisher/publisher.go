package publisher

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pseudomuto/pacman/internal/archive"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/fsutil"
	"github.com/pseudomuto/pacman/internal/types"
	"go.uber.org/fx"
)

type (
	PublisherParams struct {
		fx.In

		Packagers   []Packager `group:"publisher_packagers"`
		Persister   Persister
		Uploaders   []Uploader   `group:"publisher_uploaders"`
		VCSFetchers []VCSFetcher `group:"publisher_vcs_fetchers"`
	}

	Publisher struct {
		archivers []Packager
		persister Persister
		uploaders []Uploader
		vcs       []VCSFetcher
	}

	PublishOptions struct {
		Type        types.ArchiveType
		Storage     types.StorageType
		VCS         types.VCSType
		Repo        string
		Ref         string
		Subdir      string
		Package     string
		Description string
		Version     string
	}

	Persister interface {
		CreateArtifact(context.Context, *ent.Artifact) (*ent.Artifact, error)
	}

	Packager interface {
		Type() types.ArchiveType
		Package(context.Context, io.Writer, types.PackageOptions) error
	}

	Uploader interface {
		Type() types.StorageType
		Write(context.Context, io.Reader, string) (string, error)
	}

	VCSFetcher interface {
		Type() types.VCSType
		FetchArchive(io.Writer, string, types.VCSOptions) error
	}
)

func New(p PublisherParams) *Publisher {
	return &Publisher{
		archivers: p.Packagers,
		persister: p.Persister,
		uploaders: p.Uploaders,
		vcs:       p.VCSFetchers,
	}
}

func (p *Publisher) Publish(ctx context.Context, opts PublishOptions) error {
	packer, err := p.packager(opts.Type)
	if err != nil {
		return err
	}

	fetcher, err := p.fetcher(opts.VCS)
	if err != nil {
		return err
	}

	uploader, err := p.uploader(opts.Storage)
	if err != nil {
		return err
	}

	if err := fsutil.WithTempFile(func(tgz *os.File) error {
		// Download archive from VCS
		if err := fetcher.FetchArchive(tgz, opts.Repo, types.VCSOptions{
			Ref: opts.Ref,
			Dir: opts.Subdir,
		}); err != nil {
			return fmt.Errorf("failed to download archive from VCS: %w", err)
		}

		if _, err = tgz.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek in package: %w", err)
		}

		if err := fsutil.WithTempDir(func(dir string) error {
			// Extract archive
			if err := archive.Extract(tgz, archive.TarGz, dir, archive.StripComponents(1)); err != nil {
				return fmt.Errorf("failed to extract archive: %w", err)
			}

			// Build package archive and publish
			if err := fsutil.WithTempFile(func(pkg *os.File) error {
				if err := packer.Package(ctx, pkg, types.PackageOptions{
					Dir:     dir,
					Package: opts.Package,
					Version: opts.Version,
				}); err != nil {
					return fmt.Errorf("failed creating %s package: %w", packer.Type().String(), err)
				}

				if _, err := pkg.Seek(0, 0); err != nil {
					return fmt.Errorf("failed to seek to beginning of package: %w", err)
				}

				// Upload package
				path := fmt.Sprintf(
					"%s/%s@%s.zip",
					opts.Type,
					opts.Package,
					opts.Version,
				)

				uri, err := uploader.Write(ctx, pkg, path)
				if err != nil {
					return fmt.Errorf("failed to upload package: %w", err)
				}

				if _, err := p.persister.CreateArtifact(ctx, &ent.Artifact{
					Name:        opts.Package,
					Description: opts.Description,
					Type:        opts.Type,
					Edges: ent.ArtifactEdges{
						Versions: []*ent.ArtifactVersion{
							{
								Version: opts.Version,
								URI:     uri,
							},
						},
					},
				}); err != nil {
					return fmt.Errorf("failed to persist artifact: %w", err)
				}

				return nil
			}); err != nil {
				return fmt.Errorf("failed to write package: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to create extraction dir: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (p *Publisher) packager(t types.ArchiveType) (Packager, error) {
	for _, pkg := range p.archivers {
		if pkg.Type() == t {
			return pkg, nil
		}
	}

	return nil, fmt.Errorf("unknown packager: %d", t)
}

func (p *Publisher) fetcher(t types.VCSType) (VCSFetcher, error) {
	for _, fetcher := range p.vcs {
		if fetcher.Type() == t {
			return fetcher, nil
		}
	}

	return nil, fmt.Errorf("unknown fetcher: %d", t)
}

func (p *Publisher) uploader(t types.StorageType) (Uploader, error) {
	for _, up := range p.uploaders {
		if up.Type() == t {
			return up, nil
		}
	}

	return nil, fmt.Errorf("unknown uploader: %d", t)
}
