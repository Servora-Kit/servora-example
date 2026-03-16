package data

import (
	"context"
	"errors"

	iamconf "github.com/Servora-Kit/servora/api/gen/go/iam/conf/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/tenant"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/user"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/openfga"
)

func NewTenantRootID(ec *ent.Client, fga *openfga.Client, bizConf *iamconf.Biz, l logger.Logger) (biz.TenantRootID, error) {
	ctx := context.Background()
	t, err := ec.Tenant.Query().Where(tenant.Slug("root")).Only(ctx)
	if err != nil {
		return "", errors.New("tenant root not found: " + err.Error())
	}
	tenantID := t.ID.String()

	if fga != nil {
		seedTenantAdminFGA(ctx, ec, fga, tenantID, bizConf.GetSeed(), l)
	}

	return biz.TenantRootID(tenantID), nil
}

func seedTenant(ctx context.Context, ec *ent.Client) (string, error) {
	t, err := ec.Tenant.Query().Where(tenant.Slug("root")).Only(ctx)
	if err == nil {
		return t.ID.String(), nil
	}
	if !ent.IsNotFound(err) {
		return "", err
	}
	t, err = ec.Tenant.Create().
		SetSlug("root").
		SetName("Tenant Root").
		SetType("system").
		Save(ctx)
	if err != nil {
		return "", err
	}
	return t.ID.String(), nil
}

func seedTenantAdmin(ctx context.Context, ec *ent.Client, seed *iamconf.Biz_Seed) error {
	if seed == nil || seed.AdminEmail == "" {
		return nil
	}

	exists, err := ec.User.Query().Where(user.EmailEQ(seed.AdminEmail)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	pw, err := helpers.BcryptHash(seed.AdminPassword)
	if err != nil {
		return err
	}

	name := seed.AdminName
	if name == "" {
		name = "admin"
	}

	_, err = ec.User.Create().
		SetName(name).
		SetEmail(seed.AdminEmail).
		SetPassword(pw).
		SetRole("admin").
		Save(ctx)
	return err
}

func seedTenantAdminFGA(ctx context.Context, ec *ent.Client, fga *openfga.Client, tenantID string, seed *iamconf.Biz_Seed, l logger.Logger) {
	seedLog := logger.NewHelper(l, logger.WithModule("seed/data/iam-service"))
	if seed == nil || seed.AdminEmail == "" {
		return
	}

	u, err := ec.User.Query().Where(user.EmailEQ(seed.AdminEmail)).Only(ctx)
	if err != nil {
		return
	}

	userID := u.ID.String()
	allowed, err := fga.Check(ctx, userID, "admin", "tenant", tenantID)
	if err != nil {
		seedLog.Warnf("seed FGA check failed: %v", err)
		return
	}
	if allowed {
		return
	}

	if err := fga.WriteTuples(ctx, openfga.Tuple{
		User:     "user:" + userID,
		Relation: "admin",
		Object:   "tenant:" + tenantID,
	}); err != nil {
		seedLog.Warnf("seed tenant admin FGA tuple: %v", err)
		return
	}
	seedLog.Infof("seeded tenant admin FGA tuple for %s", seed.AdminEmail)
}
