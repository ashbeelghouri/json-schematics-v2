package schematics

import (
	"errors"
	"fmt"
)

// ValidateSchema runs every check that Check does — unknown rule/operator/
// condition names, empty targets, duplicate targets — and additionally
// verifies that every field's target actually resolves against sample once
// flattened. sample should be a representative example of the data you
// intend to validate: a test fixture, a golden payload, or real (secret-
// stripped) production data.
//
// Check alone cannot catch a typo inside a target's value. Given
// `"target": "mane"` instead of `"name"`, "mane" is a perfectly valid string,
// so Check has nothing to object to — the field just silently never matches
// anything at runtime. ValidateSchema closes that gap by matching every
// target against real data and flagging the ones that match nothing.
//
// A typo in the JSON *key* itself — `"tagret"` instead of `"target"` — never
// needs sample data to catch: encoding/json silently drops an unrecognized
// key, so Target is left as "" and Check already reports
// "field #N has an empty target". ValidateSchema surfaces that too, since it
// runs Check first.
//
// Pass ignoreTargets for fields that are legitimately allowed to be missing
// from every sample you have — an optional field with no example value in
// your fixtures, say — so they are not flagged as suspected typos.
//
// Wildcard (target containing "*") and targetRegex fields are matched the
// same way they are at validation time: if nothing in sample matches the
// pattern, the field is flagged. If that's simply because your sample
// doesn't happen to contain that shape, add the target to ignoreTargets
// rather than treating the result as a bug in the schema.
//
// ValidateSchema does not replace tests. It is meant to run in CI or a unit
// test alongside a fixture, so drift between your schema and your real data
// shape fails the build instead of failing silently in production.
func (s *Schematics) ValidateSchema(sample any, ignoreTargets ...string) error {
	var problems []string

	if err := s.Check(); err != nil {
		var se *SchemaError
		if errors.As(err, &se) {
			problems = append(problems, se.Problems...)
		} else {
			problems = append(problems, err.Error())
		}
	}

	ignore := make(map[string]bool, len(ignoreTargets))
	for _, t := range ignoreTargets {
		ignore[t] = true
	}

	flat, err := flattenSample(sample, s.separator)
	if err != nil {
		problems = append(problems, fmt.Sprintf("sample data: %v", err))
	} else {
		for i, f := range s.schema.Fields {
			if f.Target == "" || ignore[f.Target] {
				continue // empty target is already reported by Check above
			}
			if len(matchTarget(flat, f.Target, s.separator, f.TargetRegex)) == 0 {
				problems = append(problems, fmt.Sprintf(
					"field #%d: target %q does not match anything in the sample data (typo, or renamed field?)",
					i, f.Target))
			}
		}
	}

	if len(problems) > 0 {
		return &SchemaError{Problems: problems}
	}
	return nil
}

// flattenSample normalizes sample (an object, a struct, or an array of
// either) into a single flattened map. Array samples have every row merged
// together so a field exercised by only some rows is still recognized.
func flattenSample(sample any, separator string) (map[string]any, error) {
	obj, arr, err := normalize(sample)
	if err != nil {
		return nil, err
	}
	if arr == nil {
		return Flatten(obj, separator), nil
	}
	flat := map[string]any{}
	for _, row := range arr {
		for k, v := range Flatten(row, separator) {
			flat[k] = v
		}
	}
	return flat, nil
}
