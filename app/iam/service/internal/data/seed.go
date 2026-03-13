package data

import (
	"context"
	"errors"

	iamconf "github.com/Servora-Kit/servora/api/gen/go/iam/conf/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/platform"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/user"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/openfga"
)

func NewPlatformRootID(ec *ent.Client, fga *openfga.Client, bizConf *iamconf.Biz, l logger.Logger) (biz.PlatformRootID, error) {
	ctx := context.Background()
	p, err := ec.Platform.Query().Where(platform.Slug("root")).Only(ctx)
	if err != nil {
		return "", errors.New("platform root not found: " + err.Error())
	}
	platID := p.ID.String()

	if fga != nil {
		seedPlatformAdminFGA(ctx, ec, fga, platID, bizConf.GetSeed(), l)
	}

	return biz.PlatformRootID(platID), nil
}

func seedPlatform(ctx context.Context, ec *ent.Client) (string, error) {
	p, err := ec.Platform.Query().Where(platform.Slug("root")).Only(ctx)
	if err == nil {
		return p.ID.String(), nil
	}
	if !ent.IsNotFound(err) {
		return "", err
	}
	p, err = ec.Platform.Create().
		SetSlug("root").
		SetName("Platform Root").
		SetType("system").
		Save(ctx)
	if err != nil {
		return "", err
	}
	return p.ID.String(), nil
}

func seedPlatformAdmin(ctx context.Context, ec *ent.Client, seed *iamconf.Biz_Seed) error {
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

func seedPlatformAdminFGA(ctx context.Context, ec *ent.Client, fga *openfga.Client, platID string, seed *iamconf.Biz_Seed, l logger.Logger) {
	seedLog := logger.NewHelper(l, logger.WithModule("seed/data/iam-service"))
	if seed == nil || seed.AdminEmail == "" {
		return
	}

	u, err := ec.User.Query().Where(user.EmailEQ(seed.AdminEmail)).Only(ctx)
	if err != nil {
		return
	}

	userID := u.ID.String()
	allowed, err := fga.Check(ctx, userID, "admin", "platform", platID)
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
		Object:   "platform:" + platID,
	}); err != nil {
		seedLog.Warnf("seed platform admin FGA tuple: %v", err)
		return
	}
	seedLog.Infof("seeded platform admin FGA tuple for %s", seed.AdminEmail)
}
