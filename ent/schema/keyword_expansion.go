package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type KeywordExpansion struct {
	ent.Schema
}

func (KeywordExpansion) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable(),
		field.String("keyword").
			NotEmpty(),
		field.String("expansion").
			NotEmpty(),
		field.Time("created_at").
			Immutable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
	}
}

func (KeywordExpansion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("keyword"),
	}
}
