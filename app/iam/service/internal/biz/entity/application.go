package entity

import "time"

type Application struct {
	ID               string
	ClientID         string
	ClientSecretHash string
	Name             string
	RedirectURIs     []string
	Scopes           []string
	GrantTypes       []string
	ApplicationType  string
	AccessTokenType  string
	TenantID         string
	IDTokenLifetime  time.Duration
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
