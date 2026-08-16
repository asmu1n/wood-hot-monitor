package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Keyword struct {
	ent.Schema
}

func (Keyword) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable(),
		field.String("text").
			NotEmpty(),
		field.String("category").
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true),
		field.Time("created_at").
			Immutable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
		field.Time("updated_at").
			SchemaType(map[string]string{"sqlite3": "datetime"}),
	}
}

func (Keyword) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("hotspots", Hotspot.Type),
	}
}
