# Eliminate Entity Layer — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 消灭 `biz/entity` 包，让 proto message (`userpb.User` / `apppb.Application`) 作为 biz 层数据类型直通全栈。

**Architecture:** 修改 proto 定义新增完整资源 message (`User` / `Application`)，消灭旧 `UserInfo` / `ApplicationInfo`；将 biz 层接口签名从 entity 类型改为 proto 类型；将 data 层 mapper 从 `ent → entity` 改为 `ent → proto`；消灭 `service/mapper.go`；敏感字段（password、client_secret_hash）通过 repo 独立参数传递。

**Tech Stack:** Go, Protobuf/Buf, Ent ORM, Kratos, Wire

**Design doc:** `docs/plans/2026-03-21-eliminate-entity-proto-passthrough-design.md`

---

## Task 1: 修改 User proto — 新增 `User` message，重构 Request/Response

**Files:**
- Modify: `app/iam/service/api/protos/user/service/v1/user.proto`

**Step 1: 修改 proto 定义**

将 `UserInfo` 重命名为 `User`，增加 `email_verified_at`、`created_at`、`updated_at` 字段。修改所有 Request/Response 使用新 `User` message。

```proto
syntax = "proto3";

package user.service.v1;

import "buf/validate/validate.proto";
import "errors/errors.proto";
import "google/protobuf/timestamp.proto";
import "pagination/v1/pagination.proto";

option go_package = "github.com/Servora-Kit/servora/api/gen/go/user/service/v1;userpb";
option java_multiple_files = true;
option java_outer_classname = "UserProtoV1";
option java_package = "dev.servora.api.user.v1";

enum ErrorReason {
  option (errors.default_code) = 500;
  USER_NOT_FOUND = 0 [(errors.code) = 404];
  DELETE_USER_FAILED = 1 [(errors.code) = 500];
  UPDATE_USER_FAILED = 2 [(errors.code) = 500];
  SAVE_USER_FAILED = 3 [(errors.code) = 500];
  CREATE_USER_FAILED = 4 [(errors.code) = 500];
}

service UserService {
  rpc CurrentUserInfo(CurrentUserInfoRequest) returns (CurrentUserInfoResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  rpc PurgeUser(PurgeUserRequest) returns (PurgeUserResponse);
  rpc RestoreUser(RestoreUserRequest) returns (RestoreUserResponse);
}

message UserProfile {
  string name = 1;
  string given_name = 2;
  string family_name = 3;
  string nickname = 4;
  string picture = 5;
  string gender = 6;
  string birthdate = 7;
  string zoneinfo = 8;
  string locale = 9;
}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string role = 4;
  string status = 5;
  bool email_verified = 6;
  string phone = 7;
  bool phone_verified = 8;
  optional google.protobuf.Timestamp email_verified_at = 9;
  UserProfile profile = 50;
  optional google.protobuf.Timestamp created_at = 100;
  optional google.protobuf.Timestamp updated_at = 101;
}

message CurrentUserInfoRequest {}
message CurrentUserInfoResponse {
  User user = 1;
}

message GetUserRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message GetUserResponse {
  User user = 1;
}

message ListUsersRequest {
  pagination.v1.PaginationRequest pagination = 1;
}
message ListUsersResponse {
  repeated User users = 1;
  pagination.v1.PaginationResponse pagination = 2;
}

message UpdateUserRequest {
  string id = 1;
  User data = 2;
}
message UpdateUserResponse {
  User user = 1;
}

message CreateUserRequest {
  User data = 1;
  string password = 2 [(buf.validate.field).string = { min_len: 6, max_len: 64 }];
}
message CreateUserResponse {
  User user = 1;
}

message DeleteUserRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message DeleteUserResponse {
  bool success = 1;
}

message PurgeUserRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message PurgeUserResponse {
  bool success = 1;
}

message RestoreUserRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message RestoreUserResponse {
  User user = 1;
}
```

**Step 2: 运行代码生成**

```bash
cd /Users/horonlee/projects/go/servora && make api
```

Expected: 生成成功，`api/gen/go/user/service/v1/` 下的 Go 文件更新。

**Step 3: Commit**

```bash
git add app/iam/service/api/protos/user/ api/gen/go/user/
git commit -m "feat(api/proto): replace UserInfo with full User resource message"
```

---

## Task 2: 修改 Application proto — 新增 `Application` message

**Files:**
- Modify: `app/iam/service/api/protos/application/service/v1/application.proto`

**Step 1: 修改 proto 定义**

将 `ApplicationInfo` 重命名为 `Application`，修改所有 Request/Response。`CreateApplicationRequest` 改为包裹 `Application data`。

```proto
syntax = "proto3";

package application.service.v1;

import "buf/validate/validate.proto";
import "errors/errors.proto";
import "google/protobuf/timestamp.proto";
import "pagination/v1/pagination.proto";

option go_package = "github.com/Servora-Kit/servora/api/gen/go/application/service/v1;apppb";
option java_multiple_files = true;
option java_outer_classname = "ApplicationProtoV1";
option java_package = "dev.servora.api.application.v1";

enum ErrorReason {
  option (errors.default_code) = 500;
  APPLICATION_NOT_FOUND = 0 [(errors.code) = 404];
  APPLICATION_ALREADY_EXISTS = 1 [(errors.code) = 409];
  APPLICATION_CREATE_FAILED = 2 [(errors.code) = 500];
  APPLICATION_UPDATE_FAILED = 3 [(errors.code) = 500];
  APPLICATION_DELETE_FAILED = 4 [(errors.code) = 500];
  INVALID_CLIENT_SECRET = 5 [(errors.code) = 401];
}

service ApplicationService {
  rpc CreateApplication(CreateApplicationRequest) returns (CreateApplicationResponse);
  rpc GetApplication(GetApplicationRequest) returns (GetApplicationResponse);
  rpc ListApplications(ListApplicationsRequest) returns (ListApplicationsResponse);
  rpc UpdateApplication(UpdateApplicationRequest) returns (UpdateApplicationResponse);
  rpc DeleteApplication(DeleteApplicationRequest) returns (DeleteApplicationResponse);
  rpc RegenerateClientSecret(RegenerateClientSecretRequest) returns (RegenerateClientSecretResponse);
}

message Application {
  string id = 1;
  string client_id = 2;
  string name = 3;
  repeated string redirect_uris = 4;
  repeated string scopes = 5;
  repeated string grant_types = 6;
  string application_type = 7;
  string access_token_type = 8;
  string type = 9;
  int32 id_token_lifetime = 10;
  optional google.protobuf.Timestamp created_at = 100;
  optional google.protobuf.Timestamp updated_at = 101;
}

message CreateApplicationRequest {
  Application data = 1;
}
message CreateApplicationResponse {
  Application application = 1;
  string client_secret = 2;
}

message GetApplicationRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message GetApplicationResponse {
  Application application = 1;
}

message ListApplicationsRequest {
  pagination.v1.PaginationRequest pagination = 1;
}
message ListApplicationsResponse {
  repeated Application applications = 1;
  pagination.v1.PaginationResponse pagination = 2;
}

message UpdateApplicationRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
  Application data = 2;
}
message UpdateApplicationResponse {
  Application application = 1;
}

message DeleteApplicationRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message DeleteApplicationResponse {
  bool success = 1;
}

message RegenerateClientSecretRequest {
  string id = 1 [(buf.validate.field).string.uuid = true];
}
message RegenerateClientSecretResponse {
  string client_secret = 1;
}
```

**Step 2: 运行代码生成**

```bash
cd /Users/horonlee/projects/go/servora && make api
```

**Step 3: Commit**

```bash
git add app/iam/service/api/protos/application/ api/gen/go/application/
git commit -m "feat(api/proto): replace ApplicationInfo with full Application resource message"
```

---

## Task 3: 修改 biz 层 — repo 接口 + usecase 签名

**Files:**
- Modify: `app/iam/service/internal/biz/user.go`
- Modify: `app/iam/service/internal/biz/authn.go`
- Modify: `app/iam/service/internal/biz/application.go`

**Step 1: 修改 `biz/user.go`**

- 删除 `entity` import，增加 `userpb` import
- `UserRepo` 接口全部改为 proto 类型
- `UserUsecase` 方法签名改为 proto 类型
- `DeleteUser`/`PurgeUser` 参数从 `*entity.User` 改为 `string`（只用 ID）

关键变化：
- `SaveUser(context.Context, *entity.User) (*entity.User, error)` → `SaveUser(context.Context, *userpb.User, string) (*userpb.User, error)`
- `DeleteUser(context.Context, *entity.User) (*entity.User, error)` → `DeleteUser(context.Context, string) error`
- `PurgeUser(context.Context, *entity.User) (*entity.User, error)` → `PurgeUser(context.Context, string) error`
- `CreateUser` 增加 `password string` 参数
- 删除 `"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"` import（`ent.IsNotFound` 改为 data 层处理或 Kratos errors 包装）

注意：`biz/user.go` 当前 import `ent` 包来做 `ent.IsNotFound(err)` 判断。这违反分层，但属于本次重构的额外 scope。如果改动量可控可以一并修复（data 层返回 Kratos error），否则先保留。

**Step 2: 修改 `biz/authn.go`**

- 删除 `entity` import，增加 `userpb` import
- `AuthnRepo` 接口改为 proto 类型 + 新增 `GetPasswordHash`
- `SignupByEmail` 签名改为 `(ctx, username, email, password string) (*userpb.User, error)`
- `LoginByEmailPassword` 签名改为 `(ctx, email, password string) (*TokenPair, error)`
- `LoginByEmailPassword` 内部改为调用 `GetPasswordHash` 取 hash + `GetUserByID` 取 user
- `SendVerificationEmail` 参数改为 `*userpb.User`

**Step 3: 修改 `biz/application.go`**

- 删除 `entity` import，增加 `apppb` import
- `ApplicationRepo` 接口改为 proto 类型
- `Create` 签名改为 `(ctx, app *apppb.Application) (*apppb.Application, string, error)`
- `IDTokenLifetime` 不再需要 `time.Duration` 转换（proto 里已经是 `int32` 秒）

**Step 4: 验证编译**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && go build ./...
```

Expected: 编译失败（data/service 层还没改），但 biz 包本身的类型应该正确。

**Step 5: Commit**

```bash
git add app/iam/service/internal/biz/
git commit -m "refactor(app/biz): replace entity types with proto messages in biz layer"
```

---

## Task 4: 修改 data 层 mapper — ent → proto

**Files:**
- Modify: `app/iam/service/internal/data/mapper.go`

**Step 1: 重写 mapper**

将 `ent.User → entity.User` 改为 `ent.User → userpb.User`，将 `ent.Application → entity.Application` 改为 `ent.Application → apppb.Application`。

```go
package data

import (
	apppb "github.com/Servora-Kit/servora/api/gen/go/application/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/pkg/mapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var userMapper = mapper.NewForwardMapper(func(u *ent.User) *userpb.User {
	pbUser := &userpb.User{
		Id:            u.ID.String(),
		Username:      u.Username,
		Email:         u.Email,
		Role:          u.Role,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		Phone:         u.Phone,
		PhoneVerified: u.PhoneVerified,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
	if u.EmailVerifiedAt != nil {
		pbUser.EmailVerifiedAt = timestamppb.New(*u.EmailVerifiedAt)
	}
	if u.Profile != nil {
		pbUser.Profile = profileFromJSON(u.Profile)
	}
	return pbUser
})

func profileFromJSON(m map[string]interface{}) *userpb.UserProfile {
	if m == nil {
		return nil
	}
	p := &userpb.UserProfile{}
	if v, ok := m["name"].(string); ok {
		p.Name = v
	}
	if v, ok := m["given_name"].(string); ok {
		p.GivenName = v
	}
	if v, ok := m["family_name"].(string); ok {
		p.FamilyName = v
	}
	if v, ok := m["nickname"].(string); ok {
		p.Nickname = v
	}
	if v, ok := m["picture"].(string); ok {
		p.Picture = v
	}
	if v, ok := m["gender"].(string); ok {
		p.Gender = v
	}
	if v, ok := m["birthdate"].(string); ok {
		p.Birthdate = v
	}
	if v, ok := m["zoneinfo"].(string); ok {
		p.Zoneinfo = v
	}
	if v, ok := m["locale"].(string); ok {
		p.Locale = v
	}
	return p
}

var applicationMapper = mapper.NewForwardMapper(func(a *ent.Application) *apppb.Application {
	return &apppb.Application{
		Id:              a.ID.String(),
		ClientId:        a.ClientID,
		Name:            a.Name,
		RedirectUris:    a.RedirectUris,
		Scopes:          a.Scopes,
		GrantTypes:      a.GrantTypes,
		ApplicationType: a.ApplicationType,
		AccessTokenType: a.AccessTokenType,
		Type:            a.Type,
		IdTokenLifetime: int32(a.IDTokenLifetime),
		CreatedAt:       timestamppb.New(a.CreatedAt),
		UpdatedAt:       timestamppb.New(a.UpdatedAt),
	}
})
```

**Step 2: Commit**

```bash
git add app/iam/service/internal/data/mapper.go
git commit -m "refactor(app/data): rewrite mappers from ent→proto (was ent→entity)"
```

---

## Task 5: 修改 data 层 repo 实现

**Files:**
- Modify: `app/iam/service/internal/data/user.go`
- Modify: `app/iam/service/internal/data/authn.go`
- Modify: `app/iam/service/internal/data/application.go`

**Step 1: 修改 `data/user.go`**

- 删除 `entity` import，增加 `userpb` import
- 所有方法签名改为匹配新 `UserRepo` 接口
- `SaveUser` 增加 `hashedPassword string` 参数；从 `*userpb.User` 读字段时注意 proto 生成的 Go 命名（`u.Id` 而非 `u.ID`）
- `profileToJSON` 改为接收 `*userpb.UserProfile`
- `DeleteUser`/`PurgeUser` 参数从 `*entity.User` 改为 `string`

**Step 2: 修改 `data/authn.go`**

- 删除 `entity` import，增加 `userpb` import
- `SaveUser` 改为接收 `*userpb.User` + `hashedPassword string`；删除内部 bcrypt hash（hash 由 biz 层传入）
- 新增 `GetPasswordHash` 方法
- 其他方法签名对齐新接口

**Step 3: 修改 `data/application.go`**

- 删除 `entity` import，增加 `apppb` import
- `Create` 改为接收 `*apppb.Application` + `clientSecretHash string`
- `Update` 改为接收 `*apppb.Application`
- proto 生成的字段名适配（`app.ClientId`、`app.RedirectUris`、`app.IdTokenLifetime` 等）

**Step 4: 验证编译**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && go build ./...
```

Expected: data 包编译通过（service 层可能还有错误）。

**Step 5: Commit**

```bash
git add app/iam/service/internal/data/
git commit -m "refactor(app/data): update repo implementations to use proto types"
```

---

## Task 6: 修改 data 层 OIDC client

**Files:**
- Modify: `app/iam/service/internal/data/oidc_client.go`
- Modify: `app/iam/service/internal/data/oidc_client_test.go`

**Step 1: 修改 `oidc_client.go`**

- `*entity.Application` → `*apppb.Application`
- 字段访问适配：`c.app.ClientID` → `c.app.ClientId`，`c.app.RedirectURIs` → `c.app.RedirectUris`
- `IDTokenLifetime()` 从 `c.app.IDTokenLifetime`（`time.Duration`）改为 `time.Duration(c.app.IdTokenLifetime) * time.Second`

**Step 2: 修改 `oidc_client_test.go`**

- `*entity.Application` → `*apppb.Application`
- `newTestApp()` 改为构造 `*apppb.Application`
- `IDTokenLifetime` 从 `5 * time.Minute` 改为 `300`（int32 秒）
- 字段名适配

**Step 3: 运行测试**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && go test ./internal/data/ -run TestOIDC -v
```

Expected: 所有 OIDC client 测试通过。

**Step 4: Commit**

```bash
git add app/iam/service/internal/data/oidc_client.go app/iam/service/internal/data/oidc_client_test.go
git commit -m "refactor(app/data): update OIDC client to use proto Application type"
```

---

## Task 7: 修改 service 层

**Files:**
- Modify: `app/iam/service/internal/service/user.go`
- Modify: `app/iam/service/internal/service/authn.go`
- Modify: `app/iam/service/internal/service/application.go`
- Delete: `app/iam/service/internal/service/mapper.go`

**Step 1: 删除 `service/mapper.go`**

整个文件删除。entity → proto 映射不再需要。

**Step 2: 修改 `service/user.go`**

- 删除 `entity` import
- `CurrentUserInfo`：直接返回 `&userpb.CurrentUserInfoResponse{User: user}`
- `GetUser`：直接返回 `&userpb.GetUserResponse{User: user}`
- `ListUsers`：直接返回 `users` 切片
- `UpdateUser`：`s.uc.UpdateUser(ctx, callerID, req.Data)` → 直接返回
- `CreateUser`：`s.uc.CreateUser(ctx, req.Data, req.Password)` → 返回 `{User: user}`
- `DeleteUser`/`PurgeUser`：`s.uc.DeleteUser(ctx, req.Id)` → 返回 success
- `RestoreUser`：直接返回 User

**Step 3: 修改 `service/authn.go`**

- 删除 `entity` import
- `SignupByEmail`：`s.uc.SignupByEmail(ctx, req.Name, req.Email, req.Password)` → 返回 user 字段
- `LoginByEmailPassword`：`s.uc.LoginByEmailPassword(ctx, req.Email, req.Password)` → 返回 token pair

**Step 4: 修改 `service/application.go`**

- 删除 `entity` import，删除 `time` import
- `CreateApplication`：`s.uc.Create(ctx, req.Data)` → 返回 `{Application: app, ClientSecret: secret}`
- `UpdateApplication`：`s.uc.Update(ctx, req.Data)` → 返回
- 删除所有 `applicationInfoMapper` 引用

**Step 5: 验证编译**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && go build ./...
```

Expected: 全部编译通过。

**Step 6: Commit**

```bash
git add app/iam/service/internal/service/
git commit -m "refactor(app/service): simplify service layer after entity elimination"
```

---

## Task 8: 删除 entity 包 + 修改测试

**Files:**
- Delete: `app/iam/service/internal/biz/entity/user.go`
- Delete: `app/iam/service/internal/biz/entity/application.go`
- Delete: `app/iam/service/internal/biz/entity/` (directory)
- Modify: `app/iam/service/internal/biz/user_test.go`
- Modify: `app/iam/service/internal/biz/application_test.go`

**Step 1: 修改 `biz/user_test.go`**

- 删除 `entity` import，增加 `userpb` import
- `fakeUserRepo` 方法签名全部对齐新 `UserRepo` 接口
- `fakeAuthnRepo` 方法签名全部对齐新 `AuthnRepo` 接口（含 `GetPasswordHash`）
- `PurgeUser` 测试中 `&entity.User{ID: "user-1"}` 改为 `"user-1"`

**Step 2: 修改 `biz/application_test.go`**

- 删除 `entity` import，增加 `apppb` import
- `fakeApplicationRepo` 方法签名对齐新 `ApplicationRepo` 接口
- `Create` 改为 `(ctx, *apppb.Application, string) (*apppb.Application, error)`
- 测试中构造 `*apppb.Application` 替换 `*entity.Application`
- `ClientSecretHash` 不在 proto 里，测试中通过 repo 内部 map 存储 hash（fake repo 需要额外字段）

**Step 3: 删除 entity 目录**

```bash
rm -rf app/iam/service/internal/biz/entity/
```

**Step 4: 运行全部测试**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && go test ./... -v
```

Expected: 所有测试通过。

**Step 5: Commit**

```bash
git add -A app/iam/service/
git commit -m "refactor(app): delete biz/entity package, update all tests to use proto types"
```

---

## Task 9: 运行 wire + lint + 最终验证

**Files:**
- No new files

**Step 1: 运行 wire**

```bash
cd /Users/horonlee/projects/go/servora/app/iam/service && make wire
```

Expected: wire_gen.go 更新成功（如果 provider 签名变了）。

**Step 2: 运行 lint**

```bash
cd /Users/horonlee/projects/go/servora && make lint.go
```

Expected: lint 通过（可能有一些 proto 生成代码的已知 lint 忽略）。

**Step 3: 运行全部测试**

```bash
cd /Users/horonlee/projects/go/servora && make test
```

Expected: 所有测试通过。

**Step 4: 验证 TS 生成（如有）**

```bash
cd /Users/horonlee/projects/go/servora && make api-ts
```

Expected: TS 类型更新，不含 password 字段。

**Step 5: Commit**

```bash
git add -A
git commit -m "chore(app): regenerate wire + verify lint and tests after entity elimination"
```

---

## Summary

| Task | 内容 | 关键文件 |
|------|------|---------|
| 1 | User proto 重定义 | `user.proto` |
| 2 | Application proto 重定义 | `application.proto` |
| 3 | Biz 层接口 + usecase 签名 | `biz/user.go`, `biz/authn.go`, `biz/application.go` |
| 4 | Data 层 mapper 重写 | `data/mapper.go` |
| 5 | Data 层 repo 实现 | `data/user.go`, `data/authn.go`, `data/application.go` |
| 6 | OIDC client 适配 | `data/oidc_client.go`, `data/oidc_client_test.go` |
| 7 | Service 层简化 | `service/user.go`, `service/authn.go`, `service/application.go`, 删除 `service/mapper.go` |
| 8 | 删除 entity + 更新测试 | 删除 `biz/entity/`，改 `*_test.go` |
| 9 | Wire + lint + 最终验证 | 全局 |
