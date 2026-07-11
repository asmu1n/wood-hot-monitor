package schema

import (
	domain "wood-hot-monitor/internal/hotspot"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Hotspot struct {
	ent.Schema
}

func (Hotspot) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable(),
		field.String("title"),
		field.String("content"),
		field.String("url"),
		field.String("source"),
		field.String("source_id").
			Optional().
			Nillable(),
		field.Bool("is_real").
			Default(true),
		field.Int("relevance").
			Default(0),
		field.String("relevance_reason").
			Optional().
			Nillable(),
		field.Bool("keyword_mentioned").
			Optional().
			Nillable(),
		field.Enum("importance").
			GoType(domain.Importance("")).
			Default(string(domain.ImportanceLow)),
		field.String("summary").
			Optional().
			Nillable(),
		field.Int("view_count").
			Optional().
			Nillable(),
		field.Int("like_count").
			Optional().
			Nillable(),
		field.Int("retweet_count").
			Optional().
			Nillable(),
		field.Int("reply_count").
			Optional().
			Nillable(),
		field.Int("comment_count").
			Optional().
			Nillable(),
		field.Int("quote_count").
			Optional().
			Nillable(),
		field.Int("danmaku_count").
			Optional().
			Nillable(),
		field.String("author_name").
			Optional().
			Nillable(),
		field.String("author_username").
			Optional().
			Nillable(),
		field.String("author_avatar").
			Optional().
			Nillable(),
		field.Int("author_followers").
			Optional().
			Nillable(),
		field.Bool("author_verified").
			Optional().
			Nillable(),
		field.Time("published_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
		field.Time("created_at").
			Immutable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
		field.String("keyword_id").
			Optional().
			Nillable(),
		field.Bool("is_notified").
			Default(false),
		field.Time("notified_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{"sqlite3": "datetime"}),
		field.Bool("is_read").
			Default(false),
	}
}

func (Hotspot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("keyword", Keyword.Type).
			Ref("hotspots").
			Field("keyword_id").
			Unique(),
	}
}

func (Hotspot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("url", "source").
			Unique(),
		index.Fields("is_notified", "is_read"),
	}
}
