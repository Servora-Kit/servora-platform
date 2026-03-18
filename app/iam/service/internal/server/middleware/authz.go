package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	authzpb "github.com/Servora-Kit/servora/api/gen/go/authz/service/v1"
	iamv1 "github.com/Servora-Kit/servora/api/gen/go/iam/service/v1"
	"github.com/Servora-Kit/servora/pkg/actor"
	"github.com/Servora-Kit/servora/pkg/openfga"
	"github.com/Servora-Kit/servora/pkg/redis"
)

// AuthzOption configures the Authz middleware.
type AuthzOption func(*authzConfig)

type authzConfig struct {
	fga      *openfga.Client
	redis    *redis.Client
	cacheTTL time.Duration
	rules    map[string]iamv1.AuthzRuleEntry
}

func WithFGAClient(c *openfga.Client) AuthzOption {
	return func(cfg *authzConfig) { cfg.fga = c }
}

// WithAuthzRules sets the operation→rule mapping directly from generated code.
func WithAuthzRules(rules map[string]iamv1.AuthzRuleEntry) AuthzOption {
	return func(cfg *authzConfig) { cfg.rules = rules }
}

func WithAuthzCache(rdb *redis.Client, ttl time.Duration) AuthzOption {
	return func(cfg *authzConfig) {
		cfg.redis = rdb
		cfg.cacheTTL = ttl
	}
}

// Authz creates a Kratos middleware that performs authorization checks
// using OpenFGA based on proto-declared rules.
//
// Behavior:
//   - AUTHZ_MODE_NONE: skip authorization
//   - AUTHZ_MODE_ORGANIZATION: check relation on organization:{id}
//   - AUTHZ_MODE_OBJECT: check relation on {object_type}:{id}
//   - No rule found (fail-closed): deny
//   - OpenFGA unavailable (fail-closed): 503
func Authz(opts ...AuthzOption) middleware.Middleware {
	cfg := &authzConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			operation := tr.Operation()
			rule, found := cfg.rules[operation]
			if !found {
				return nil, errors.Forbidden("AUTHZ_NO_RULE",
					fmt.Sprintf("no authorization rule for operation %s", operation))
			}

			if rule.Mode == authzpb.AuthzMode_AUTHZ_MODE_NONE {
				return handler(ctx, req)
			}

			a, ok := actor.FromContext(ctx)
			if !ok || a.Type() != actor.TypeUser {
				return nil, errors.Forbidden("AUTHZ_DENIED", "authentication required")
			}
			userID := a.ID()

			if cfg.fga == nil {
				return nil, errors.ServiceUnavailable("AUTHZ_UNAVAILABLE", "authorization service not available")
			}

			objectType, objectID, err := resolveObject(rule, req, a)
			if err != nil {
				return nil, errors.BadRequest("AUTHZ_BAD_REQUEST",
					fmt.Sprintf("cannot resolve authorization target: %v", err))
			}

			relation := relationToFGA(rule.Relation)
			ttl := cfg.cacheTTL
			if ttl == 0 {
				ttl = openfga.DefaultCheckCacheTTL
			}
			allowed, err := cfg.fga.CachedCheck(ctx, cfg.redis, ttl,
				userID, relation, objectType, objectID)
			if err != nil {
				return nil, errors.ServiceUnavailable("AUTHZ_CHECK_FAILED",
					fmt.Sprintf("authorization check failed: %v", err))
			}
			if !allowed {
				return nil, errors.Forbidden("AUTHZ_DENIED", "insufficient permissions")
			}

			return handler(ctx, req)
		}
	}
}

func resolveObject(rule iamv1.AuthzRuleEntry, req any, a actor.Actor) (objectType, objectID string, err error) {
	switch rule.Mode {
	case authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION:
		objectType = "organization"
		if rule.IDField == "" {
			objectID, err = scopeFromActor(a, "OrganizationID")
		} else {
			objectID, err = extractProtoField(req, rule.IDField)
		}
	case authzpb.AuthzMode_AUTHZ_MODE_OBJECT:
		objectType = objectTypeToFGA(rule.ObjectType)
		if objectType == "tenant" && rule.IDField == "" {
			objectID, err = scopeFromActor(a, "TenantID")
		} else {
			objectID, err = extractProtoField(req, rule.IDField)
		}
	default:
		err = fmt.Errorf("unsupported authz mode: %v", rule.Mode)
	}
	return
}

func scopeFromActor(a actor.Actor, field string) (string, error) {
	ua, ok := a.(*actor.UserActor)
	if !ok {
		return "", fmt.Errorf("actor is not a UserActor")
	}
	switch field {
	case "TenantID":
		if id := ua.TenantID(); id != "" {
			return id, nil
		}
		return "", fmt.Errorf("missing X-Tenant-ID header")
	case "OrganizationID":
		if id := ua.OrganizationID(); id != "" {
			return id, nil
		}
		return "", fmt.Errorf("missing X-Organization-ID header")
	default:
		return "", fmt.Errorf("unknown scope field: %s", field)
	}
}

func extractProtoField(req any, fieldName string) (string, error) {
	if fieldName == "" {
		return "", fmt.Errorf("id_field not specified")
	}

	msg, ok := req.(proto.Message)
	if !ok {
		return "", fmt.Errorf("request is not a proto message")
	}

	md := msg.ProtoReflect().Descriptor()
	fd := md.Fields().ByName(protoreflect.Name(fieldName))
	if fd == nil {
		return "", fmt.Errorf("field %q not found in %s", fieldName, md.FullName())
	}

	val := msg.ProtoReflect().Get(fd)
	s := val.String()
	if s == "" {
		return "", fmt.Errorf("field %q is empty", fieldName)
	}
	return s, nil
}

func relationToFGA(r authzpb.Relation) string {
	s := strings.TrimPrefix(r.String(), "RELATION_")
	return strings.ToLower(s)
}

func objectTypeToFGA(ot authzpb.ObjectType) string {
	s := strings.TrimPrefix(ot.String(), "OBJECT_TYPE_")
	return strings.ToLower(s)
}
