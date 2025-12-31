package data

import (
	"context"
	"fmt"

	"github.com/pseudomuto/pacman/internal/ent"
)

type (
	Repo struct {
		db *ent.Client
	}
)

func NewRepo(db *ent.Client) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateArtifact(ctx context.Context, af *ent.Artifact) (*ent.Artifact, error) {
	return WithTx(ctx, r.db, func(tx *ent.Tx) (*ent.Artifact, error) {
		artifact, err := tx.Artifact.Create().
			SetName(af.Name).
			SetDescription(af.Description).
			SetType(af.Type).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create artifact: %w", err)
		}

		versions := make([]*ent.ArtifactVersionCreate, len(af.Edges.Versions))
		for i, v := range af.Edges.Versions {
			versions[i] = tx.ArtifactVersion.Create().
				SetArtifact(artifact).
				SetVersion(v.Version).
				SetURI(v.URI)
		}

		if len(versions) > 0 {
			vers, err := tx.ArtifactVersion.
				CreateBulk(versions...).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create versions for artifact: %w", err)
			}

			artifact.Edges.Versions = vers
			for _, v := range vers {
				v.Edges.Artifact = artifact
			}
		}

		return artifact, nil
	})
}
