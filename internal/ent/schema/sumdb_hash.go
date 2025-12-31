package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SumDBHash struct {
	ent.Schema
}

func (SumDBHash) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (SumDBHash) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("index"),
		field.Bytes("hash"),
	}
}

func (SumDBHash) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tree"),
		index.Fields("index").Edges("tree").Unique(),
	}
}

func (SumDBHash) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("tree", SumDBTree.Type).
			Ref("hashes").
			Required().
			Immutable().
			Unique(),
	}
}
