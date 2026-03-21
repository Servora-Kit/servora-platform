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

	err := ApplyPlan(plan, m, presets, hooks)
	require.NoError(t, err)

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

	err := ApplyPlan(plan, m, presets, hooks)
	require.NoError(t, err)
}

func TestMapperPlan_ApplyMissingPreset(t *testing.T) {
	plan := &MapperPlan{
		Presets: []string{"nonexistent_preset"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := ApplyPlan(plan, m, presets, hooks)
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

	err := ApplyPlan(plan, m, presets, hooks)
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

func TestMapperPlan_FieldConverters(t *testing.T) {
	plan := &MapperPlan{
		FieldConverters: map[string]ConverterKind{
			"CreatedAt": ConverterTimestampTime,
			"UpdatedAt": ConverterTimestampTime,
		},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := ApplyPlan(plan, m, presets, hooks)
	require.NoError(t, err)
}

func TestMapperPlan_InvalidFieldConverter(t *testing.T) {
	plan := &MapperPlan{
		FieldConverters: map[string]ConverterKind{
			"SomeField": "NONEXISTENT_KIND",
		},
	}

	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := plan.Validate(presets, hooks)
	require.Error(t, err)
	require.Contains(t, err.Error(), "NONEXISTENT_KIND")
	require.Contains(t, err.Error(), "SomeField")
}

func TestMapperPlan_IgnoredFields(t *testing.T) {
	plan := &MapperPlan{
		Presets:       []string{"common_proto_entity"},
		IgnoredFields: []string{"Password", "Secret"},
	}

	m := NewCopierMapper[domainUser, entLikeUser]()
	presets := DefaultPresets()
	hooks := NewHookRegistry()

	err := ApplyPlan(plan, m, presets, hooks)
	require.NoError(t, err)
}
