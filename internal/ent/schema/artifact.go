package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/pseudomuto/pacman/internal/types"
)

// Artifact holds the schema definition for the Artifact entity.
type Artifact struct {
	ent.Schema
}

func (Artifact) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields of the Artifact.
func (Artifact) Fields() []ent.Field {
	return []ent.Field{
		field.
			String("name").
			MaxLen(300).
			Unique(),
		field.Text("description"),
		field.Enum("type").GoType(types.ArchiveType(-1)),
	}
}

// Edges of the Artifact.
func (Artifact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("versions", ArtifactVersion.Type).
			StorageKey(edge.Column("artifact_id")),
	}
}
