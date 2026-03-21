package mapper

import (
	"fmt"
	"strings"
)

// MapperPlan is a declarative specification for how to configure a CopierMapper.
// Generated code produces MapperPlan values; runtime applies them.
type MapperPlan struct {
	Presets      []string
	FieldMapping map[string]string
	CustomHooks  []string
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

// ApplyPlan configures a CopierMapper according to the given plan.
// Go does not support type parameters on methods, so this is a package-level function.
func ApplyPlan[P any, E any](plan *MapperPlan, m *CopierMapper[P, E], presets *PresetRegistry, hooks *HookRegistry) error {
	if err := plan.Validate(presets, hooks); err != nil {
		return err
	}

	if len(plan.Presets) > 0 {
		cs, err := presets.Collect(plan.Presets...)
		if err != nil {
			return err
		}
		m.AppendConverters(cs)
	}

	if len(plan.FieldMapping) > 0 {
		m.WithFieldMapping(plan.FieldMapping)
	}

	for _, name := range plan.CustomHooks {
		cs := hooks.MustGet(name)
		m.AppendConverters(cs)
	}

	return nil
}
