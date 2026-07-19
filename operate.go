package schematics

import (
	"context"
	"fmt"
)

// Operate applies each field's operator chain to the matching values in data
// and returns the transformed document. data may be a single object or an array
// of objects. Validators are not run.
func (s *Schematics) Operate(data any) (any, error) {
	return s.OperateCtx(context.Background(), data)
}

// OperateCtx is Operate with a caller-supplied context.Context.
func (s *Schematics) OperateCtx(ctx context.Context, data any) (any, error) {
	if err := s.ensureChecked(); err != nil {
		return nil, err
	}
	obj, arr, err := normalize(data)
	if err != nil {
		return nil, err
	}
	if arr != nil {
		out := make([]map[string]any, 0, len(arr))
		for _, o := range arr {
			res, err := s.operateObject(ctx, o)
			if err != nil {
				return nil, err
			}
			out = append(out, res)
		}
		return out, nil
	}
	return s.operateObject(ctx, obj)
}

func (s *Schematics) operateObject(ctx context.Context, obj map[string]any) (map[string]any, error) {
	flat := Flatten(obj, s.separator)
	db := s.buildDB(flat)

	for _, f := range s.schema.Fields {
		if len(f.Operate) == 0 {
			continue
		}
		view := &FieldView{Target: f.Target, Name: f.Name, Type: f.Type, Required: f.Required, Tags: f.Tags}
		matches := matchTarget(flat, f.Target, s.separator, f.TargetRegex)
		for key, val := range matches {
			view.Provided = true
			cctx := &Context{
				Ctx:       ctx,
				DB:        db,
				Locale:    s.locale,
				Separator: s.separator,
				Flat:      flat,
				Field:     view,
			}
			newVal, err := s.applyOperators(f, val, cctx)
			if err != nil {
				return nil, fmt.Errorf("operating on %q: %w", key, err)
			}
			flat[key] = newVal
		}
	}
	return Deflate(flat, s.separator), nil
}

func (s *Schematics) applyOperators(f Field, val any, cctx *Context) (any, error) {
	for _, ref := range f.Operate {
		op := s.operators[ref.Op] // presence guaranteed by Check
		args := ref.Args
		if args == nil {
			args = Args{}
		}
		next, err := op(val, args, cctx)
		if err != nil {
			return nil, fmt.Errorf("operator %q: %w", ref.Op, err)
		}
		val = next
	}
	return val, nil
}
