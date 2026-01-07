package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/pseudomuto/pacman/internal/types"
)

type Archive struct {
	ent.Schema
}

func (Archive) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (Archive) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").GoType(types.ArchiveType(-1)),
		field.String("coordinate").MaxLen(200),
		field.JSON("assets", []AssetURL{}),
	}
}

func (Archive) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("type", "coordinate").Unique(),
	}
}
