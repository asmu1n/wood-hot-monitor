package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable(),
		field.String("type").
			NotEmpty(),
		field.String("title"),
		field.String("content"),
		field.Bool("is_read").
			Default(false),
		field.String("hotspot_id").
			Optional().
			Nillable(),
		field.Time("created_at").
			Immutable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
	}
}

func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("hotspot", Hotspot.Type).
			Field("hotspot_id").
			Unique(),
	}
}
