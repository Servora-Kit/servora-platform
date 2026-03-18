package entity

import "time"

type Tenant struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Domain      string
	Kind        string // "business" | "personal"
	Status      string // "active" | "disabled"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TenantMember struct {
	ID        string
	TenantID  string
	UserID    string
	UserName  string
	UserEmail string
	Role      string // "owner" | "admin" | "member"
	Status    string // "active" | "invited"
	JoinedAt  *time.Time
	CreatedAt time.Time
}
