package data

import (
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/pkg/mapper"
)

var userMapper = mapper.NewForwardMapper(func(u *ent.User) *entity.User {
	return &entity.User{
		ID:       u.ID.String(),
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Role:     u.Role,
	}
})

var orgMapper = mapper.NewForwardMapper(func(o *ent.Organization) *entity.Organization {
	e := &entity.Organization{
		ID:         o.ID.String(),
		PlatformID: o.PlatformID.String(),
		Name:       o.Name,
		Slug:       o.Slug,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
	if o.DisplayName != nil {
		e.DisplayName = *o.DisplayName
	}
	return e
})

var projectMapper = mapper.NewForwardMapper(func(p *ent.Project) *entity.Project {
	e := &entity.Project{
		ID:             p.ID.String(),
		OrganizationID: p.OrganizationID.String(),
		Name:           p.Name,
		Slug:           p.Slug,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
	if p.Description != nil {
		e.Description = *p.Description
	}
	return e
})
