package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SumDBRecord struct {
	ent.Schema
}

func (SumDBRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (SumDBRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("record_id"),
		field.String("path").MaxLen(200),
		field.String("version").MaxLen(20),
		field.Bytes("data"),
	}
}

func (SumDBRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tree"),
		index.Edges("tree").Fields("record_id").Unique(),
		index.Edges("tree").Fields("path", "version").Unique(),
	}
}

func (SumDBRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("tree", SumDBTree.Type).
			Ref("records").
			Required().
			Immutable().
			Unique(),
	}
}
