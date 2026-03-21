# 设计文档：消灭 entity 层，proto message 直通 biz

**日期：** 2026-03-21
**状态：** 已实施完成 (2026-03-21)

---

## 1. 背景与动机

当前 Servora IAM 服务存在一个 `biz/entity` 包，定义了 `entity.User`、`entity.Application` 等贫血模型。这些类型：

- 与 proto message（`UserInfo`、`ApplicationInfo`）高度重复；
- 与 ent entity（`ent.User`、`ent.Application`）也高度重复；
- 导致每个请求经历三层映射：proto → entity → ent；
- 增加了 service/data 层的手写映射代码量；
- 作为贫血模型不承载任何业务逻辑，不解决任何实际问题。

本设计消灭 entity 层，让 proto message 作为 biz 层的数据类型直通全栈。

### 参考实现

- **go-wind-admin**：proto 资源 message 全栈流转，无中间 entity 层，repo 直接接收/返回 proto；
- **kratos-cli（sql2proto）**：从数据库/SQL 生成包含所有列的资源 proto message。

---

## 2. 核心决策

| 决策点 | 结论 |
|--------|------|
| `biz/entity/` 包 | 消灭 |
| 资源 proto | 定义完整 `User` / `Application`，不含敏感字段 |
| 敏感字段 | repo 方法用独立参数传递，biz 用 `GetPasswordHash` 取 hash 自己比对 |
| biz 层 | 保留，签名改为接收/返回 proto message |
| service 层 | 保留，职责为 Request/Response 适配 + transport 前置操作 |
| data 层 | 纯 CRUD，mapper 从 ent → proto |
| `service/mapper.go` | 消灭（service 层不再需要 entity → proto 映射） |
| `data/mapper.go` | 改为 ent → proto 映射 |
| Request 包裹模式 | `CreateUserRequest { User data = 1 }` |

---

## 3. 新数据流

### 旧流程（三层模型）

```text
proto Request
  → service 层手写拆装 → entity.User
  → biz 层操作 entity.User
  → data 层 mapper: entity.User ↔ ent.User
  → service 层 mapper: entity.User → proto Response
```

### 新流程（两层模型）

```text
proto Request
  → service 层取 req.Data（*userpb.User）
  → biz 层操作 *userpb.User
  → data 层 mapper: ent.User → *userpb.User
  → service 层直接包 response
```

---

## 4. Proto 资源模型

### 4.1 User

```proto
message User {
  string id = 1;
  string username = 2;
  string email = 3;
  // 不含 password — 敏感字段通过 repo 独立参数传递
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
```

消灭现有 `UserInfo`，统一为 `User`。

### 4.2 Application

```proto
message Application {
  string id = 1;
  string client_id = 2;
  // 不含 client_secret_hash — 敏感字段通过 repo 独立参数传递
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
```

消灭现有 `ApplicationInfo`，统一为 `Application`。

### 4.3 Request/Response 包裹模式

写操作采用 go-wind-admin 的 data 包裹模式：

```proto
message CreateUserRequest {
  User data = 1;
  string password = 2;  // 敏感字段放在 request wrapper 上
}
message CreateUserResponse { User user = 1; }

message UpdateUserRequest {
  string id = 1;
  User data = 2;
}
message UpdateUserResponse { User user = 1; }
```

读操作直接返回 `User`：

```proto
message GetUserResponse { User user = 1; }
message ListUsersResponse {
  repeated User users = 1;
  pagination.v1.PaginationResponse pagination = 2;
}
message CurrentUserInfoResponse { User user = 1; }
```

### 4.4 安全性保证

敏感字段（password、client_secret_hash）不在资源 proto 中定义。这意味着：

- 无论 mapper 怎么通用化，proto 类型上就不存在这些字段，物理上不可能泄露；
- 通用 codegen 可以安全地做全字段映射；
- TS 前端生成类型中不会出现敏感字段。

---

## 5. 敏感字段处理

### 5.1 原则

data 层只做纯 CRUD，不做密码校验等业务逻辑。敏感字段通过 repo 方法的独立参数传递。

### 5.2 Password

```go
// data 层 — 纯存取
type AuthnRepo interface {
    SaveUser(ctx context.Context, user *userpb.User, hashedPassword string) (*userpb.User, error)
    GetPasswordHash(ctx context.Context, email string) (userID string, hash string, err error)
    UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
    // ...
}

// biz 层 — 业务逻辑
func (uc *AuthnUsecase) LoginByEmailPassword(ctx context.Context, email, password string) (*TokenPair, error) {
    userID, hash, err := uc.repo.GetPasswordHash(ctx, email)
    if err != nil { /* ... */ }
    if !helpers.BcryptCheck(password, hash) {
        return nil, authnpb.ErrorIncorrectPassword("invalid email or password")
    }
    user, err := uc.repo.GetUserByID(ctx, userID)
    // ... 生成 token
}
```

### 5.3 ClientSecretHash

```go
type ApplicationRepo interface {
    Create(ctx context.Context, app *apppb.Application, clientSecretHash string) (*apppb.Application, error)
    UpdateClientSecretHash(ctx context.Context, id string, hash string) error
    GetClientSecretHash(ctx context.Context, clientID string) (string, error)
    // ...
}
```

---

## 6. 全栈接口签名

### 6.1 Biz 层 — Repo 接口

**UserRepo**（由 biz 定义，data 实现）：

```go
type UserRepo interface {
    SaveUser(ctx context.Context, user *userpb.User, hashedPassword string) (*userpb.User, error)
    GetUserById(ctx context.Context, id string) (*userpb.User, error)
    DeleteUser(ctx context.Context, id string) error
    PurgeUser(ctx context.Context, id string) error
    PurgeCascade(ctx context.Context, id string) error
    RestoreUser(ctx context.Context, id string) (*userpb.User, error)
    GetUserByIdIncludingDeleted(ctx context.Context, id string) (*userpb.User, error)
    UpdateUser(ctx context.Context, user *userpb.User) (*userpb.User, error)
    ListUsers(ctx context.Context, page, pageSize int32) ([]*userpb.User, int64, error)
}
```

**AuthnRepo**：

```go
type AuthnRepo interface {
    SaveUser(ctx context.Context, user *userpb.User, hashedPassword string) (*userpb.User, error)
    GetUserByEmail(ctx context.Context, email string) (*userpb.User, error)
    GetUserByUserName(ctx context.Context, username string) (*userpb.User, error)
    GetUserByID(ctx context.Context, id string) (*userpb.User, error)
    GetPasswordHash(ctx context.Context, email string) (userID string, hash string, err error)
    UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
    UpdateEmailVerified(ctx context.Context, userID string, verified bool) error
    TokenStore
}
```

**ApplicationRepo**：

```go
type ApplicationRepo interface {
    Create(ctx context.Context, app *apppb.Application, clientSecretHash string) (*apppb.Application, error)
    GetByID(ctx context.Context, id string) (*apppb.Application, error)
    GetByClientID(ctx context.Context, clientID string) (*apppb.Application, error)
    List(ctx context.Context, page, pageSize int32) ([]*apppb.Application, int64, error)
    Update(ctx context.Context, app *apppb.Application) (*apppb.Application, error)
    Delete(ctx context.Context, id string) error
    UpdateClientSecretHash(ctx context.Context, id string, hash string) error
    GetClientSecretHash(ctx context.Context, clientID string) (string, error)
}
```

### 6.2 Biz 层 — Usecase 方法

**UserUsecase**：

```go
func (uc *UserUsecase) CurrentUserInfo(ctx context.Context, callerID string) (*userpb.User, error)
func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*userpb.User, error)
func (uc *UserUsecase) UpdateUser(ctx context.Context, callerID string, user *userpb.User) (*userpb.User, error)
func (uc *UserUsecase) CreateUser(ctx context.Context, user *userpb.User, password string) (*userpb.User, error)
func (uc *UserUsecase) ListUsers(ctx context.Context, page, pageSize int32) ([]*userpb.User, int64, error)
func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) (bool, error)
func (uc *UserUsecase) PurgeUser(ctx context.Context, id string) (bool, error)
func (uc *UserUsecase) RestoreUser(ctx context.Context, id string) (*userpb.User, error)
```

**AuthnUsecase**：

```go
func (uc *AuthnUsecase) SignupByEmail(ctx context.Context, username, email, password string) (*userpb.User, error)
func (uc *AuthnUsecase) LoginByEmailPassword(ctx context.Context, email, password string) (*TokenPair, error)
func (uc *AuthnUsecase) SendVerificationEmail(ctx context.Context, user *userpb.User) error
```

### 6.3 Service 层

保留，职责明确：

- Request/Response 包装与适配
- 从 context 提取 actor/caller 身份
- 参数校验（password confirm、captcha）
- 分页请求解析与响应构建
- 未来：FieldMask 处理、请求清洗、rate limiting 等

### 6.4 Data 层 Mapper

从 `ent → entity` 改为 `ent → proto`：

```go
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
        Profile:       profileFromJSON(u.Profile),
    }
    if u.EmailVerifiedAt != nil {
        pbUser.EmailVerifiedAt = timestamppb.New(*u.EmailVerifiedAt)
    }
    return pbUser
})
```

---

## 7. 消灭清单

| 删除项 | 原因 |
|--------|------|
| `biz/entity/user.go` | 被 `userpb.User` 替代 |
| `biz/entity/application.go` | 被 `apppb.Application` 替代 |
| `biz/entity/` 目录 | 整个消灭 |
| `service/mapper.go` | service 层不再需要 entity → proto 映射 |

## 8. 修改清单

| 文件 | 变化 |
|------|------|
| `user.proto` | 新增 `User` message，消灭 `UserInfo`，改造 Request/Response |
| `application.proto` | 新增 `Application` message，消灭 `ApplicationInfo`，改造 Request/Response |
| `biz/user.go` | `UserRepo` 接口 + `UserUsecase` 方法签名全改为 proto 类型 |
| `biz/authn.go` | `AuthnRepo` 接口 + `AuthnUsecase` 方法签名改为 proto 类型 + 独立参数 |
| `biz/application.go` | `ApplicationRepo` 接口 + `ApplicationUsecase` 签名全改为 proto 类型 |
| `data/mapper.go` | 从 ent → entity 改为 ent → proto |
| `data/user.go` | 所有 repo 方法改为使用 proto 类型 |
| `data/authn.go` | 同上 + 新增 `GetPasswordHash` |
| `data/application.go` | 同上 |
| `data/oidc_client.go` | `*entity.Application` → `*apppb.Application`，字段访问适配 |
| `data/oidc_storage.go` | entity 引用改为 proto |
| `service/user.go` | 简化为取 `req.Data` 调 biz |
| `service/authn.go` | entity 构造改为独立参数 |
| `service/application.go` | entity 构造改为取 `req.Data` |
| `biz/user_test.go` | entity 改为 proto |
| `biz/application_test.go` | entity 改为 proto |
| `data/oidc_client_test.go` | entity 改为 proto |

---

## 9. 与 mapper codegen 设计的关系

本设计是 `2026-03-21-servora-mapper-proto-codegen-design.md` 的**前置条件**：

- 消灭 entity 后，mapper 只需做 `ent ↔ proto` 一层映射；
- proto 中不含敏感字段，通用 mapper/codegen 可以安全地做全字段映射；
- proto annotation + codegen 生成的 mapper 将直接替换当前手写的 `data/mapper.go`。

mapper codegen 设计文档需要相应更新：
- 删除"三层模型"相关讨论；
- 将默认映射关系更新为 `resource proto <-> ent entity`（已一致）；
- Phase 0 中"透传边界"部分已被本设计解决。
