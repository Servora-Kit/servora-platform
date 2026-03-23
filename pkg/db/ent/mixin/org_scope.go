package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/google/uuid"
)

type OrgScopeMixin struct {
	mixin.Schema
}

func (OrgScopeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("organization_id", uuid.UUID{}),
	}
}

func (OrgScopeMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id"),
	}
}
