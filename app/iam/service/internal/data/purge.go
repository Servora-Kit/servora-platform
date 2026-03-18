package data

import (
	"context"

	"github.com/google/uuid"

	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/organizationmember"
)

// purgeOrganizationInTx deletes an Organization and all its dependent
// resources (OrganizationMember) within the caller's existing transaction.
// The ent.Client must come from a transaction context (via Data.Ent(txCtx)).
func purgeOrganizationInTx(ctx context.Context, c *ent.Client, orgID uuid.UUID) error {
	if _, err := c.OrganizationMember.Delete().
		Where(organizationmember.OrganizationIDEQ(orgID)).
		Exec(ctx); err != nil {
		return err
	}

	return c.Organization.DeleteOneID(orgID).Exec(ctx)
}
