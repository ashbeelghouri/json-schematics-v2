package schematics

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Validate checks data against the loaded schema. data may be a single object
// (map or struct) or an array of objects. It returns nil when the data is
// valid, a *ValidationErrors when it is not, or a *SchemaError when the schema
// itself references something unregistered.
func (s *Schematics) Validate(data any) error {
	return s.ValidateCtx(context.Background(), data)
}

// ValidateCtx is Validate with a caller-supplied context.Context, made
// available to every rule via Context.Ctx.
func (s *Schematics) ValidateCtx(ctx context.Context, data any) error {
	if err := s.ensureChecked(); err != nil {
		return err
	}
	obj, arr, err := normalize(data)
	if err != nil {
		return err
	}
	errs := &ValidationErrors{}
	if arr != nil {
		s.validateArray(ctx, arr, errs)
	} else {
		s.validateObject(ctx, obj, "", errs)
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// normalize converts arbitrary input into either a single object or a slice of
// objects by round-tripping through JSON.
func normalize(data any) (map[string]any, []map[string]any, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("data is not valid JSON: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err == nil && obj != nil {
		return obj, nil, nil
	}
	var arr []map[string]any
	if err := json.Unmarshal(b, &arr); err == nil {
		return nil, arr, nil
	}
	return nil, nil, fmt.Errorf("data must be an object or an array of objects")
}

func (s *Schematics) validateArray(ctx context.Context, arr []map[string]any, errs *ValidationErrors) {
	for i, obj := range arr {
		rowID := itoa(i)
		if s.arrayIDKey != "" {
			flat := Flatten(obj, s.separator)
			if v, ok := flat[s.arrayIDKey]; ok {
				rowID = toString(v)
			}
		}
		s.validateObject(ctx, obj, rowID, errs)
	}
}

func (s *Schematics) validateObject(ctx context.Context, obj map[string]any, rowID string, errs *ValidationErrors) {
	flat := Flatten(obj, s.separator)
	db := s.buildDB(flat)

	// First pass: figure out which fields are present so dependencies and
	// conditions can be evaluated.
	matchesByIndex := make([]map[string]any, len(s.schema.Fields))
	provided := map[string]bool{}
	for i, f := range s.schema.Fields {
		m := matchTarget(flat, f.Target, s.separator, f.TargetRegex)
		matchesByIndex[i] = m
		if len(m) > 0 {
			provided[f.Target] = true
		}
	}

	for i, f := range s.schema.Fields {
		matches := matchesByIndex[i]
		view := &FieldView{
			Target:   f.Target,
			Name:     f.Name,
			Type:     f.Type,
			Required: f.Required,
			Provided: len(matches) > 0,
			Tags:     f.Tags,
		}
		cctx := &Context{
			Ctx:       ctx,
			DB:        db,
			Locale:    s.locale,
			Separator: s.separator,
			Flat:      flat,
			RowID:     rowID,
			Field:     view,
		}

		if !s.conditionsPass(f, cctx) {
			continue
		}

		if missing := missingDeps(f, provided); len(missing) > 0 {
			// Only surface a dependency error when the field itself matters.
			if f.Required || len(matches) > 0 {
				e := newError(f.Target, "dependsOn", nil)
				e.RowID = rowID
				e.fallback = fmt.Sprintf("missing dependencies: %s", strings.Join(missing, ", "))
				errs.Add(e)
			}
			continue
		}

		if f.Required && len(matches) == 0 {
			e := newError(f.Target, "required", nil)
			e.RowID = rowID
			e.fallback = "this field is required"
			errs.Add(e)
			continue
		}

		if len(matches) == 0 {
			continue // optional and absent
		}

		for key, val := range matches {
			if ve := s.runRules(f, key, val, cctx); ve != nil {
				errs.Add(ve)
			}
		}
	}
}

func (s *Schematics) conditionsPass(f Field, cctx *Context) bool {
	for _, w := range f.When {
		cond, ok := s.conditions[w.Condition]
		if !ok {
			// Check() guarantees this cannot happen; be conservative and treat
			// an unknown condition as not satisfied.
			return false
		}
		args := w.Args
		if args == nil {
			args = Args{}
		}
		res := cond(args, cctx)
		if w.Negate {
			res = !res
		}
		if !res {
			return false
		}
	}
	return true
}

func missingDeps(f Field, provided map[string]bool) []string {
	var missing []string
	for _, dep := range f.DependsOn {
		if !provided[dep] {
			missing = append(missing, dep)
		}
	}
	return missing
}

// runRules evaluates a field's validators against one matched value and returns
// the first failure, or nil.
func (s *Schematics) runRules(f Field, key string, val any, cctx *Context) *ValidationError {
	for _, ref := range f.Validate {
		rule := s.rules[ref.Rule] // presence guaranteed by Check
		args := ref.Args
		if args == nil {
			args = Args{}
		}
		if err := rule(val, args, cctx); err != nil {
			e := newError(key, ref.Rule, val)
			e.RowID = cctx.RowID
			e.fallback = err.Error()
			if ref.Message != "" {
				e.fallback = ref.Message
				e.setMessage(s.locale, ref.Message)
			}
			for loc, m := range ref.Messages {
				e.setMessage(loc, m)
			}
			return e
		}
	}
	return nil
}

// buildDB assembles the read-only DB handed to rules: the base DB plus any
// fields flagged addToDB.
func (s *Schematics) buildDB(flat map[string]any) map[string]any {
	db := make(map[string]any, len(s.db))
	for k, v := range s.db {
		db[k] = v
	}
	for _, f := range s.schema.Fields {
		if !f.AddToDB {
			continue
		}
		m := matchTarget(flat, f.Target, s.separator, f.TargetRegex)
		switch len(m) {
		case 0:
			// nothing to add
		case 1:
			for _, v := range m {
				db[f.Target] = v
			}
		default:
			vals := make([]any, 0, len(m))
			for _, v := range m {
				vals = append(vals, v)
			}
			db[f.Target] = vals
		}
	}
	return db
}
