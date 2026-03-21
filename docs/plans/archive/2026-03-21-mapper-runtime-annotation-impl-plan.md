# Mapper Runtime & Proto Annotation — Implementation Plan (Batch A)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 重构 `pkg/mapper` 运行时（增强版 `CopierMapper`、preset 体系、custom hook registry、MapperPlan/apply/validate）并定义 `servora.mapper` proto annotation，为后续 `protoc-gen-servora-mapper` 插件打下基础。

**Architecture:** `pkg/mapper` 提供泛型运行时——`CopierMapper[P,E]` 基于 `jinzhu/copier` 做反射映射，通过 `MapperPlan` 结构声明 preset/field mapping/converter/custom hook，`Apply` 一次性装配，`Validate` 在初始化期检查完整性。Proto annotation 定义在共享 `api/protos/mapper/v1/mapper.proto`，声明 message 级与 field 级规则，供后续 protoc 插件读取。

**Tech Stack:** Go 1.24, Protobuf, jinzhu/copier v0.4.0, Buf v2, testify

**Design Doc:** `docs/plans/2026-03-21-servora-mapper-proto-codegen-design.md` — Sections 7 & 8

---

## Task 1: 重构 `CopierMapper[P, E]` 核心类型

**Files:**
- Modify: `pkg/mapper/copier_proto.go` (rename & rewrite)
- Delete: `pkg/mapper/copier_db.go` (merged into unified CopierMapper)
- Modify: `pkg/mapper/mapper_test.go` (update tests)

**Step 1: Write failing tests for new CopierMapper**

Add to `pkg/mapper/mapper_test.go`:

```go
// ---------- CopierMapper (new unified) ----------

func TestCopierMapper_ToProto(t *testing.T) {
	m := NewCopierMapper[entLikeUser, domainUser]()
	m.AppendConverters(AllBuiltinConverters())

	src := &entLikeUser{ID: 7, Name: "alice", Email: "alice@example.com", Password: "hashed", Role: "admin"}
	dst, err := m.ToProto(src)
	require.NoError(t, err)
	require.NotNil(t, dst)
	require.Equal(t, src.ID, dst.ID)
	require.Equal(t, src.Name, dst.Name)
}

func TestCopierMapper_ToEntity(t *testing.T) {
	m := NewCopierMapper[entLikeUser, domainUser]()
	m.AppendConverters(AllBuiltinConverters())

	src := &domainUser{ID: 3, Name: "bob", Email: "bob@example.com", Role: "user"}
	dst, err := m.ToEntity(src)
	require.NoError(t, err)
	require.NotNil(t, dst)
	require.Equal(t, src.ID, dst.ID)
}

func TestCopierMapper_ErrorReturn(t *testing.T) {
	m := NewCopierMapper[entLikeUser, domainUser]()
	// nil input
	dst, err := m.ToProto(nil)
	require.NoError(t, err)
	require.Nil(t, dst)
}

func TestCopierMapper_MustToProto(t *testing.T) {
	m := NewCopierMapper[entLikeUser, domainUser]()
	m.AppendConverters(AllBuiltinConverters())

	src := &entLikeUser{ID: 1, Name: "test"}
	dst := m.MustToProto(src)
	require.NotNil(t, dst)
	require.Equal(t, int64(1), dst.ID)
}

func TestCopierMapper_ListConversions(t *testing.T) {
	m := NewCopierMapper[entLikeUser, domainUser]()
	m.AppendConverters(AllBuiltinConverters())

	entities := []*entLikeUser{{ID: 1, Name: "a"}, nil, {ID: 2, Name: "b"}}
	protos, err := m.ToProtoList(entities)
	require.NoError(t, err)
	require.Len(t, protos, 2)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd pkg/mapper && go test ./... -v -run TestCopierMapper`
Expected: compilation errors — `NewCopierMapper` not found

**Step 3: Implement `CopierMapper[P, E]`**

Rewrite `pkg/mapper/copier_proto.go` to contain the new unified `CopierMapper`:

```go
package mapper

import "github.com/jinzhu/copier"

// CopierMapper is a reflection-based bidirectional mapper between
// Proto/domain type P and Entity/storage type E.
// Type parameter order: P = proto/domain side, E = entity/storage side.
type CopierMapper[P any, E any] struct {
	converters   []copier.TypeConverter
	fieldMapping []copier.FieldNameMapping
	options      copier.Option
}

func NewCopierMapper[P any, E any]() *CopierMapper[P, E] {
	return &CopierMapper[P, E]{
		converters: make([]copier.TypeConverter, 0),
		options: copier.Option{
			IgnoreEmpty: false,
			DeepCopy:    true,
		},
	}
}

func (m *CopierMapper[P, E]) AppendConverter(c copier.TypeConverter) *CopierMapper[P, E] {
	m.converters = append(m.converters, c)
	return m
}

func (m *CopierMapper[P, E]) AppendConverters(cs []copier.TypeConverter) *CopierMapper[P, E] {
	m.converters = append(m.converters, cs...)
	return m
}

func (m *CopierMapper[P, E]) WithFieldMapping(mapping map[string]string) *CopierMapper[P, E] {
	for src, dst := range mapping {
		m.fieldMapping = append(m.fieldMapping,
			copier.FieldNameMapping{SrcType: new(E), DstType: new(P), Mapping: map[string]string{src: dst}},
			copier.FieldNameMapping{SrcType: new(P), DstType: new(E), Mapping: map[string]string{dst: src}},
		)
	}
	return m
}

func (m *CopierMapper[P, E]) buildOption() copier.Option {
	opt := m.options
	opt.Converters = m.converters
	if len(m.fieldMapping) > 0 {
		opt.FieldNameMapping = m.fieldMapping
	}
	return opt
}

// ToProto converts entity E to proto P. Returns (nil, nil) when input is nil.
func (m *CopierMapper[P, E]) ToProto(entity *E) (*P, error) {
	if entity == nil {
		return nil, nil
	}
	var p P
	if err := copier.CopyWithOption(&p, entity, m.buildOption()); err != nil {
		return nil, err
	}
	return &p, nil
}

// ToEntity converts proto P to entity E. Returns (nil, nil) when input is nil.
func (m *CopierMapper[P, E]) ToEntity(proto *P) (*E, error) {
	if proto == nil {
		return nil, nil
	}
	var e E
	if err := copier.CopyWithOption(&e, proto, m.buildOption()); err != nil {
		return nil, err
	}
	return &e, nil
}

// MustToProto converts entity to proto, panics on error.
func (m *CopierMapper[P, E]) MustToProto(entity *E) *P {
	p, err := m.ToProto(entity)
	if err != nil {
		panic("mapper: ToProto: " + err.Error())
	}
	return p
}

// MustToEntity converts proto to entity, panics on error.
func (m *CopierMapper[P, E]) MustToEntity(proto *P) *E {
	e, err := m.ToEntity(proto)
	if err != nil {
		panic("mapper: ToEntity: " + err.Error())
	}
	return e
}

func (m *CopierMapper[P, E]) ToProtoList(entities []*E) ([]*P, error) {
	if len(entities) == 0 {
		return nil, nil
	}
	result := make([]*P, 0, len(entities))
	for _, e := range entities {
		p, err := m.ToProto(e)
		if err != nil {
			return nil, err
		}
		if p != nil {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *CopierMapper[P, E]) ToEntityList(protos []*P) ([]*E, error) {
	if len(protos) == 0 {
		return nil, nil
	}
	result := make([]*E, 0, len(protos))
	for _, p := range protos {
		e, err := m.ToEntity(p)
		if err != nil {
			return nil, err
		}
		if e != nil {
			result = append(result, e)
		}
	}
	return result, nil
}
```

**Step 4: Delete `copier_db.go`**

`CopierMapper` now unifies proto and DB mapper. Delete `pkg/mapper/copier_db.go`.

Remove old `CopierProtoMapper` type alias/tests — update `TestCopierDBMapperWithEntLikeStruct` and `TestCopierDBMapperListWithNilItems` to use `CopierMapper`.

**Step 5: Run tests**

Run: `cd pkg/mapper && go test ./... -v`
Expected: all pass

**Step 6: Commit**

```bash
git add pkg/mapper/
git commit -m "refactor(pkg/mapper): unify CopierProtoMapper and CopierDBMapper into CopierMapper[P,E]"
```

---

## Task 2: Preset 体系

**Files:**
- Create: `pkg/mapper/preset.go`
- Create: `pkg/mapper/preset_test.go`

**Step 1: Write failing tests**

```go
// preset_test.go
package mapper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPresetRegistry_GetBuiltin(t *testing.T) {
	r := NewPresetRegistry()
	r.RegisterDefaults()

	cs, ok := r.Get("proto_time")
	require.True(t, ok)
	require.NotEmpty(t, cs)
}

func TestPresetRegistry_GetUnknown(t *testing.T) {
	r := NewPresetRegistry()
	_, ok := r.Get("nonexistent")
	require.False(t, ok)
}

func TestPresetRegistry_Collect(t *testing.T) {
	r := NewPresetRegistry()
	r.RegisterDefaults()

	cs, err := r.Collect("proto_time", "pointer")
	require.NoError(t, err)
	require.NotEmpty(t, cs)
}

func TestPresetRegistry_CollectUnknown(t *testing.T) {
	r := NewPresetRegistry()
	_, err := r.Collect("nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent")
}

func TestPresetRegistry_CommonProtoEntity(t *testing.T) {
	r := NewPresetRegistry()
	r.RegisterDefaults()

	cs, ok := r.Get("common_proto_entity")
	require.True(t, ok)
	require.NotEmpty(t, cs)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd pkg/mapper && go test ./... -v -run TestPresetRegistry`
Expected: compilation errors

**Step 3: Implement preset registry**

```go
// preset.go
package mapper

import (
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
)

// PresetRegistry manages named groups of converters.
type PresetRegistry struct {
	presets map[string]func() []copier.TypeConverter
}

func NewPresetRegistry() *PresetRegistry {
	return &PresetRegistry{presets: make(map[string]func() []copier.TypeConverter)}
}

func (r *PresetRegistry) Register(name string, factory func() []copier.TypeConverter) {
	r.presets[name] = factory
}

func (r *PresetRegistry) Get(name string) ([]copier.TypeConverter, bool) {
	f, ok := r.presets[name]
	if !ok {
		return nil, false
	}
	return f(), true
}

// Collect gathers converters from multiple presets; returns error listing any missing.
func (r *PresetRegistry) Collect(names ...string) ([]copier.TypeConverter, error) {
	var missing []string
	var result []copier.TypeConverter
	for _, n := range names {
		cs, ok := r.Get(n)
		if !ok {
			missing = append(missing, n)
			continue
		}
		result = append(result, cs...)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("mapper: unknown presets: %s", strings.Join(missing, ", "))
	}
	return result, nil
}

// RegisterDefaults registers all built-in presets.
func (r *PresetRegistry) RegisterDefaults() {
	r.Register("proto_time", NewTimestamppbConverterPair)
	r.Register("time_ptr", NewTimeConverterPair)
	r.Register("pointer", func() []copier.TypeConverter {
		var cs []copier.TypeConverter
		cs = append(cs, NewStringPointerConverterPair()...)
		cs = append(cs, NewInt64PointerConverterPair()...)
		return cs
	})
	r.Register("common_proto_entity", func() []copier.TypeConverter {
		var cs []copier.TypeConverter
		cs = append(cs, NewTimestamppbConverterPair()...)
		cs = append(cs, NewTimeConverterPair()...)
		cs = append(cs, NewStringPointerConverterPair()...)
		cs = append(cs, NewInt64PointerConverterPair()...)
		return cs
	})
}

// DefaultPresets returns a PresetRegistry with all built-in presets registered.
func DefaultPresets() *PresetRegistry {
	r := NewPresetRegistry()
	r.RegisterDefaults()
	return r
}
```

**Step 4: Run tests**

Run: `cd pkg/mapper && go test ./... -v -run TestPresetRegistry`
Expected: all pass

**Step 5: Commit**

```bash
git add pkg/mapper/preset.go pkg/mapper/preset_test.go
git commit -m "feat(pkg/mapper): add preset registry for named converter groups"
```

---

## Task 3: Custom hook registry

**Files:**
- Create: `pkg/mapper/hook.go`
- Create: `pkg/mapper/hook_test.go`

**Step 1: Write failing tests**

```go
// hook_test.go
package mapper

import (
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/require"
)

func TestHookRegistry_RegisterAndGet(t *testing.T) {
	r := NewHookRegistry()
	r.Register("user_profile", copier.TypeConverter{
		SrcType: "", DstType: "",
		Fn: func(src any) (any, error) { return src, nil },
	})

	cs, ok := r.Get("user_profile")
	require.True(t, ok)
	require.Len(t, cs, 1)
}

func TestHookRegistry_GetMissing(t *testing.T) {
	r := NewHookRegistry()
	_, ok := r.Get("nonexistent")
	require.False(t, ok)
}

func TestHookRegistry_MustGet_Panics(t *testing.T) {
	r := NewHookRegistry()
	require.Panics(t, func() { r.MustGet("nonexistent") })
}

func TestHookRegistry_CheckMissing(t *testing.T) {
	r := NewHookRegistry()
	r.Register("a", copier.TypeConverter{})

	err := r.CheckRequired("a", "b", "c")
	require.Error(t, err)
	require.Contains(t, err.Error(), "b")
	require.Contains(t, err.Error(), "c")
}

func TestHookRegistry_CheckAllPresent(t *testing.T) {
	r := NewHookRegistry()
	r.Register("a", copier.TypeConverter{})
	r.Register("b", copier.TypeConverter{})

	err := r.CheckRequired("a", "b")
	require.NoError(t, err)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd pkg/mapper && go test ./... -v -run TestHookRegistry`
Expected: compilation errors

**Step 3: Implement hook registry**

```go
// hook.go
package mapper

import (
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
)

// HookRegistry manages named custom converter hooks.
// Repos register concrete implementations; generated code references hooks by name.
type HookRegistry struct {
	hooks map[string][]copier.TypeConverter
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[string][]copier.TypeConverter)}
}

func (r *HookRegistry) Register(name string, converters ...copier.TypeConverter) {
	r.hooks[name] = append(r.hooks[name], converters...)
}

func (r *HookRegistry) Get(name string) ([]copier.TypeConverter, bool) {
	cs, ok := r.hooks[name]
	return cs, ok
}

func (r *HookRegistry) MustGet(name string) []copier.TypeConverter {
	cs, ok := r.hooks[name]
	if !ok {
		panic(fmt.Sprintf("mapper: hook %q not registered", name))
	}
	return cs
}

// CheckRequired verifies all required hook names are registered.
func (r *HookRegistry) CheckRequired(names ...string) error {
	var missing []string
	for _, n := range names {
		if _, ok := r.hooks[n]; !ok {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("mapper: missing hooks: %s", strings.Join(missing, ", "))
	}
	return nil
}
```

**Step 4: Run tests**

Run: `cd pkg/mapper && go test ./... -v -run TestHookRegistry`
Expected: all pass

**Step 5: Commit**

```bash
git add pkg/mapper/hook.go pkg/mapper/hook_test.go
git commit -m "feat(pkg/mapper): add custom hook registry for named converter hooks"
```

---

## Task 4: MapperPlan + Apply + Validate

**Files:**
- Create: `pkg/mapper/plan.go`
- Create: `pkg/mapper/plan_test.go`

**Step 1: Write failing tests**

```go
// plan_test.go
package mapper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapperPlan_ApplyToMapper(t *testing.T) {
	plan := &MapperPlan{
		Presets:      []string{"common_proto_entity"},
		FieldMapping: map[string]string{"UserName": "Name"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := plan.Apply(m, presets, hooks)
	require.NoError(t, err)

	// Verify mapping works after apply
	src := &entLikeUser{ID: 1, Name: "alice"}
	dst := m.MustToProto(src)
	require.Equal(t, int64(1), dst.ID)
}

func TestMapperPlan_ApplyWithCustomHook(t *testing.T) {
	plan := &MapperPlan{
		Presets:     []string{"proto_time"},
		CustomHooks: []string{"test_hook"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()
	hooks.Register("test_hook", AllBuiltinConverters()...)

	err := plan.Apply(m, presets, hooks)
	require.NoError(t, err)
}

func TestMapperPlan_ApplyMissingPreset(t *testing.T) {
	plan := &MapperPlan{
		Presets: []string{"nonexistent_preset"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := plan.Apply(m, presets, hooks)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent_preset")
}

func TestMapperPlan_ApplyMissingHook(t *testing.T) {
	plan := &MapperPlan{
		CustomHooks: []string{"missing_hook"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := plan.Apply(m, presets, hooks)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing_hook")
}

func TestMapperPlan_Validate(t *testing.T) {
	plan := &MapperPlan{
		Presets:     []string{"proto_time", "pointer"},
		CustomHooks: []string{"hook_a"},
	}

	presets := DefaultPresets()
	hooks := NewHookRegistry()
	hooks.Register("hook_a")

	err := plan.Validate(presets, hooks)
	require.NoError(t, err)
}

func TestMapperPlan_ValidateFails(t *testing.T) {
	plan := &MapperPlan{
		Presets:     []string{"bad_preset"},
		CustomHooks: []string{"bad_hook"},
	}

	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := plan.Validate(presets, hooks)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bad_preset")
	require.Contains(t, err.Error(), "bad_hook")
}
```

**Step 2: Run tests to verify they fail**

Run: `cd pkg/mapper && go test ./... -v -run TestMapperPlan`
Expected: compilation errors

**Step 3: Implement MapperPlan**

```go
// plan.go
package mapper

import (
	"fmt"
	"strings"
)

// MapperPlan is a declarative specification for how to configure a CopierMapper.
// Generated code produces MapperPlan values; runtime applies them.
type MapperPlan struct {
	Presets      []string          // preset names to load
	FieldMapping map[string]string // entity field -> proto field renames
	CustomHooks  []string          // required custom hook names
}

// Validate checks that all referenced presets and hooks exist without modifying any mapper.
func (p *MapperPlan) Validate(presets *PresetRegistry, hooks *HookRegistry) error {
	var errs []string

	for _, name := range p.Presets {
		if _, ok := presets.Get(name); !ok {
			errs = append(errs, fmt.Sprintf("unknown preset %q", name))
		}
	}
	for _, name := range p.CustomHooks {
		if _, ok := hooks.Get(name); !ok {
			errs = append(errs, fmt.Sprintf("missing hook %q", name))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("mapper plan validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Apply configures a CopierMapper according to this plan.
// It loads preset converters, applies field mappings, loads custom hook converters,
// and validates everything is present.
func (p *MapperPlan) Apply[P any, E any](m *CopierMapper[P, E], presets *PresetRegistry, hooks *HookRegistry) error {
	if err := p.Validate(presets, hooks); err != nil {
		return err
	}

	// Load preset converters
	if len(p.Presets) > 0 {
		cs, err := presets.Collect(p.Presets...)
		if err != nil {
			return err
		}
		m.AppendConverters(cs)
	}

	// Apply field mappings
	if len(p.FieldMapping) > 0 {
		m.WithFieldMapping(p.FieldMapping)
	}

	// Load custom hook converters
	for _, name := range p.CustomHooks {
		cs := hooks.MustGet(name)
		m.AppendConverters(cs)
	}

	return nil
}
```

> **Note:** Go 1.24 does not support type parameters on methods. `Apply` must be a top-level function:
>
> ```go
> func ApplyPlan[P any, E any](plan *MapperPlan, m *CopierMapper[P, E], presets *PresetRegistry, hooks *HookRegistry) error
> ```
>
> The implementing engineer should adjust the API surface accordingly: `plan.Validate(presets, hooks)` stays as a method, but `ApplyPlan(plan, m, presets, hooks)` becomes a package-level function. Tests should be updated to match.

**Step 4: Run tests**

Run: `cd pkg/mapper && go test ./... -v -run TestMapperPlan`
Expected: all pass

**Step 5: Run full test suite**

Run: `cd pkg/mapper && go test ./... -v`
Expected: all tests pass (old + new)

**Step 6: Commit**

```bash
git add pkg/mapper/plan.go pkg/mapper/plan_test.go
git commit -m "feat(pkg/mapper): add MapperPlan declarative configuration and ApplyPlan"
```

---

## Task 5: 定义 `servora.mapper` proto annotation

**Files:**
- Create: `api/protos/mapper/v1/mapper.proto`

**Step 1: Write the proto file**

```protobuf
syntax = "proto3";

package mapper.v1;

import "google/protobuf/descriptor.proto";

option go_package = "github.com/Servora-Kit/servora/api/gen/go/mapper/v1;mapperpb";

// ConverterKind enumerates built-in converter types that can be referenced in field annotations.
enum ConverterKind {
  CONVERTER_UNSPECIFIED = 0;
  // time.Time <-> *timestamppb.Timestamp
  TIMESTAMP_TIME = 1;
  // time.Time <-> *time.Time
  TIME_PTR = 2;
  // string <-> *string
  STRING_PTR = 3;
  // int64 <-> *int64
  INT64_PTR = 4;
  // protobuf enum int32 <-> entity string
  ENUM_STRING = 5;
}

// MapperMessageRule is a message-level annotation controlling mapper codegen.
message MapperMessageRule {
  // Whether this message participates in mapper code generation.
  bool enabled = 1;
  // Preset names to apply (e.g. "proto_time", "pointer", "common_proto_entity").
  repeated string presets = 2;
}

// MapperFieldRule is a field-level annotation controlling individual field mapping.
message MapperFieldRule {
  // Rename the target field in the entity.
  string rename = 1;
  // Use a built-in converter for this field.
  ConverterKind converter = 2;
  // Reference a custom hook by name (registered at runtime in repo).
  string custom = 3;
  // If true, this field is excluded from mapping.
  bool ignore = 4;
}

extend google.protobuf.MessageOptions {
  // mapper declares mapper code generation rules for this message.
  MapperMessageRule mapper = 50200;
}

extend google.protobuf.FieldOptions {
  // mapper_field declares per-field mapping rules.
  MapperFieldRule mapper_field = 50201;
}
```

> **Extension number 选择说明：** authz 使用了 50100；mapper 使用 50200 和 50201 避免冲突。与 Servora 自定义扩展号段 5xxxx 保持一致。

**Step 2: Verify proto compiles**

Run: `buf lint`
Expected: 0 errors

**Step 3: Generate Go code**

Run: `make api`
Expected: `api/gen/go/mapper/v1/mapper.pb.go` generated successfully

**Step 4: Commit**

```bash
git add api/protos/mapper/v1/mapper.proto api/gen/go/mapper/v1/
git commit -m "feat(api/proto): define servora.mapper message and field annotations"
```

---

## Task 6: 在 `User` proto 上添加 mapper annotation 示例

**Files:**
- Modify: `app/iam/service/api/protos/user/service/v1/user.proto`

**Step 1: Add mapper annotations to User message**

Import `mapper/v1/mapper.proto` and annotate the `User` message:

```protobuf
import "mapper/v1/mapper.proto";

message User {
  option (mapper.v1.mapper) = {
    enabled: true,
    presets: ["common_proto_entity"]
  };

  string id = 1 [(mapper.v1.mapper_field) = { rename: "ID", converter: STRING_PTR }];
  string username = 2;
  string email = 3;
  string role = 4;
  string status = 5;
  bool email_verified = 6;
  string phone = 7;
  bool phone_verified = 8;
  optional google.protobuf.Timestamp email_verified_at = 9 [(mapper.v1.mapper_field) = { converter: TIMESTAMP_TIME }];
  UserProfile profile = 50 [(mapper.v1.mapper_field) = { custom: "user_profile" }];
  optional google.protobuf.Timestamp created_at = 100 [(mapper.v1.mapper_field) = { converter: TIMESTAMP_TIME }];
  optional google.protobuf.Timestamp updated_at = 101 [(mapper.v1.mapper_field) = { converter: TIMESTAMP_TIME }];
}
```

> **注意：** 此步仅为语义声明，当前无 protoc 插件消费这些注解，但需确保编译通过、生成代码包含 extension 信息。

**Step 2: Regenerate**

Run: `make api`
Expected: 编译通过，`user.pb.go` 包含 mapper annotation 信息

**Step 3: Verify full build**

Run: `go build ./...` (from workspace root)
Expected: 0 errors

**Step 4: Commit**

```bash
git add app/iam/service/api/protos/user/service/v1/user.proto api/gen/go/
git commit -m "feat(api/proto): annotate User message with servora.mapper rules"
```

---

## Task 7: 更新 IAM data/mapper.go 使用新 CopierMapper（可选验证）

**Files:**
- Modify: `app/iam/service/internal/data/mapper.go`

此 task 是可选的 smoke-test：将 `applicationMapper` 从手写 `ForwardMapper` 切换为 `CopierMapper` + preset，验证新 runtime 在真实代码中可用。`userMapper` 因为有 `profileFromJSON` custom hook 逻辑，暂时保留手写。

**Step 1: Update applicationMapper**

```go
var applicationMapper = func() *mapper.CopierMapper[apppb.Application, ent.Application] {
	m := mapper.NewCopierMapper[apppb.Application, ent.Application]()
	m.AppendConverters(mapper.AllBuiltinConverters())
	// ent.Application.ID is uuid.UUID, proto is string — needs custom converter
	m.AppendConverters(mapper.NewGenericConverterPair[uuid.UUID, string](
		func(id uuid.UUID) (string, error) { return id.String(), nil },
		func(s string) (uuid.UUID, error) { return uuid.Parse(s) },
	))
	return m
}()
```

**Step 2: Update repo code using applicationMapper**

Existing call sites use `applicationMapper.Map(entApp)`. New CopierMapper uses `applicationMapper.MustToProto(entApp)`. Update all call sites in `app/iam/service/internal/data/application.go`.

**Step 3: Run IAM tests**

Run: `cd app/iam/service && go test ./... -v`
Expected: all pass

**Step 4: Run full lint**

Run: `cd app/iam/service && golangci-lint run ./...`
Expected: 0 issues

**Step 5: Commit**

```bash
git add app/iam/service/internal/data/
git commit -m "refactor(app/iam): migrate applicationMapper to CopierMapper runtime"
```

---

## Task 8: 更新设计文档 & 最终验证

**Files:**
- Modify: `docs/plans/2026-03-21-servora-mapper-proto-codegen-design.md`

**Step 1: Update design doc Phase 1 & 2 status**

Mark Phase 1 (16.2) and Phase 2 (16.3) as completed with implementation notes.

**Step 2: Run full verification**

```bash
go build ./...
cd pkg/mapper && go test ./... -v
cd app/iam/service && go test ./... -v
cd app/iam/service && golangci-lint run ./...
buf lint
make api
```

Expected: all pass, 0 errors, 0 lint issues

**Step 3: Commit**

```bash
git add docs/plans/
git commit -m "docs(plans): mark mapper runtime (Phase 1) and proto annotation (Phase 2) as completed"
```
