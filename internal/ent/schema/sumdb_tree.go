package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/pseudomuto/pacman/internal/crypto"
)

type SumDBTree struct {
	ent.Schema
}

func (SumDBTree) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (SumDBTree) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(200).Unique(),
		field.String("signer_key").MaxLen(100).GoType(crypto.Secret("")),
		field.String("verifier_key").MaxLen(100),
	}
}

func (SumDBTree) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("hashes", SumDBHash.Type).
			StorageKey(edge.Column("tree_id")),
		edge.To("records", SumDBRecord.Type).
			StorageKey(edge.Column("tree_id")),
	}
}
