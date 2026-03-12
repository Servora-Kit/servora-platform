package data

import (
	"context"
	"errors"
	"strings"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/platform"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/user"
	entdrv "github.com/Servora-Kit/servora/pkg/ent"
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/openfga"
	"github.com/Servora-Kit/servora/pkg/redis"
	"github.com/Servora-Kit/servora/pkg/transport/client"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var ProviderSet = wire.NewSet(registry.NewDiscovery, NewDBClient, NewPlatformRootID, NewRedis, NewData, NewAuthRepo, NewUserRepo, NewTestRepo, NewOrganizationRepo, NewProjectRepo)

type Data struct {
	entClient *ent.Client
	log       *logger.Helper
	client    client.Client
	redis     *redis.Client
}

func NewData(entClient *ent.Client, c *conf.Data, l logger.Logger, client client.Client, redisClient *redis.Client) (*Data, func(), error) {
	_ = c
	cleanup := func() {
		logger.NewHelper(l).Info("closing the data resources")
		if err := entClient.Close(); err != nil {
			logger.NewHelper(l).Warnf("failed to close ent client: %v", err)
		}
	}
	return &Data{
		entClient: entClient,
		log:       logger.NewHelper(l, logger.WithModule("core/data/iam-service")),
		client:    client,
		redis:     redisClient,
	}, cleanup, nil
}

func NewDBClient(cfg *conf.Data, app *conf.App, l logger.Logger) (*ent.Client, error) {
	driver, err := entdrv.NewDriver(cfg)
	if err != nil {
		return nil, err
	}

	opts := []ent.Option{
		ent.Driver(driver),
		ent.Log(logger.EntLogFuncFrom(l, "ent/data/iam-service")),
	}
	if strings.EqualFold(app.GetEnv(), "dev") {
		opts = append(opts, ent.Debug())
	}

	ec := ent.NewClient(opts...)

	ctx := context.Background()
	if err := ec.Schema.Create(ctx); err != nil {
		return nil, errors.New("ent auto-migrate: " + err.Error())
	}

	if _, err := seedPlatform(ctx, ec); err != nil {
		return nil, errors.New("seed platform: " + err.Error())
	}

	if err := seedPlatformAdmin(ctx, ec, app.GetSeed()); err != nil {
		seedLog := logger.NewHelper(l, logger.WithModule("seed/data/iam-service"))
		seedLog.Warnf("seed platform admin: %v", err)
	}

	return ec, nil
}

func NewPlatformRootID(ec *ent.Client, fga *openfga.Client, app *conf.App, l logger.Logger) (biz.PlatformRootID, error) {
	ctx := context.Background()
	p, err := ec.Platform.Query().Where(platform.Slug("root")).Only(ctx)
	if err != nil {
		return "", errors.New("platform root not found: " + err.Error())
	}
	platID := p.ID.String()

	if fga != nil {
		seedPlatformAdminFGA(ctx, ec, fga, platID, app.GetSeed(), l)
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

func seedPlatformAdmin(ctx context.Context, ec *ent.Client, seed *conf.App_Seed) error {
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

func seedPlatformAdminFGA(ctx context.Context, ec *ent.Client, fga *openfga.Client, platID string, seed *conf.App_Seed, l logger.Logger) {
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

func NewRedis(cfg *conf.Data, l logger.Logger) (*redis.Client, func(), error) {
	redisConfig := redis.NewConfigFromProto(cfg.Redis)
	if redisConfig == nil {
		return nil, nil, errors.New("redis configuration is required")
	}

	return redis.NewClient(redisConfig, logger.With(l, logger.WithModule("redis/data/iam-service")))
}
