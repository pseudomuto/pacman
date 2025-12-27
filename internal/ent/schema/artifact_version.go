package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ArtifactVersion holds the schema definition for the ArtifactVersion entity.
type ArtifactVersion struct {
	ent.Schema
}

func (ArtifactVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields of the ArtifactVersion.
func (ArtifactVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").MaxLen(50),
		field.String("uri").MaxLen(1000),
	}
}

// Edges of the ArtifactVersion.
func (ArtifactVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("artifact", Artifact.Type).
			Ref("versions").
			Required().
			Immutable().
			Unique().
			Comment("The artifact this version belongs to"),
	}
}

func (ArtifactVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version").Edges("artifact").Unique(),
	}
}
