package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/pseudomuto/pacman/internal/types"
)

type Asset struct {
	ent.Schema
}

func (Asset) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (Asset) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").GoType(types.AssetType(-1)),
		field.String("uri").MaxLen(2048),
	}
}

func (Asset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sumdb_records", SumDBRecord.Type).Ref("assets"),
	}
}
