package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(newUUIDv7),
		field.String("slug").MaxLen(64).Unique(),
		field.String("name").MaxLen(128),
		field.String("type").MaxLen(32).Default("system"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organizations", Organization.Type),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tenants"},
	}
}
