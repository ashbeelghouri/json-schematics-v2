package schematics

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSchemaCatchesValueTypo(t *testing.T) {
	// "target": "mane" is a typo for "name" — syntactically valid JSON, so
	// Check alone has nothing to flag. ValidateSchema needs sample data to
	// notice it matches nothing.
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"target":"mane","required":true}]}`)); err != nil {
		t.Fatal(err)
	}
	sample := map[string]any{"name": "Ada", "email": "ada@example.com"}

	err := s.ValidateSchema(sample)
	var se *SchemaError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SchemaError, got %v", err)
	}
	if !strings.Contains(se.Error(), `"mane"`) {
		t.Errorf("expected error to mention the typoed target, got: %v", se)
	}
}

func TestValidateSchemaCatchesKeyTypoViaCheck(t *testing.T) {
	// "tagret" is a typo for "target" in the JSON key itself. encoding/json
	// drops the unknown key, Target ends up "", and Check — which
	// ValidateSchema runs first — already reports it. No sample data needed.
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"tagret":"name","required":true}]}`)); err != nil {
		t.Fatal(err)
	}
	err := s.ValidateSchema(map[string]any{"name": "Ada"})
	var se *SchemaError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SchemaError, got %v", err)
	}
	if !strings.Contains(se.Error(), "empty target") {
		t.Errorf("expected an empty-target problem, got: %v", se)
	}
}

func TestValidateSchemaHappyPath(t *testing.T) {
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[
		{"target":"name","required":true},
		{"target":"email","validate":[{"rule":"email"}]}]}`)); err != nil {
		t.Fatal(err)
	}
	sample := map[string]any{"name": "Ada", "email": "ada@example.com"}
	if err := s.ValidateSchema(sample); err != nil {
		t.Fatalf("expected a clean schema, got %v", err)
	}
}

func TestValidateSchemaIgnoreTargets(t *testing.T) {
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[
		{"target":"name","required":true},
		{"target":"middleName"}]}`)); err != nil {
		t.Fatal(err)
	}
	sample := map[string]any{"name": "Ada"} // no fixture happens to carry a middle name

	if err := s.ValidateSchema(sample); err == nil {
		t.Fatal("expected middleName to be flagged without ignoreTargets")
	}
	if err := s.ValidateSchema(sample, "middleName"); err != nil {
		t.Fatalf("expected ignoreTargets to suppress it, got %v", err)
	}
}

func TestValidateSchemaArraySampleMergesRows(t *testing.T) {
	// Different rows may exercise different optional fields; ValidateSchema
	// merges every row before matching targets.
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"target":"a"},{"target":"b"}]}`)); err != nil {
		t.Fatal(err)
	}
	sample := []map[string]any{
		{"a": 1},
		{"b": 2},
	}
	if err := s.ValidateSchema(sample); err != nil {
		t.Fatalf("expected a clean schema, got %v", err)
	}
}

func TestValidateSchemaStillCatchesUnknownRule(t *testing.T) {
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"target":"x","validate":[{"rule":"noSuchRule"}]}]}`)); err != nil {
		t.Fatal(err)
	}
	err := s.ValidateSchema(map[string]any{"x": "y"})
	var se *SchemaError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SchemaError, got %v", err)
	}
}

func TestValidateSchemaBadSample(t *testing.T) {
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"target":"x"}]}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateSchema(func() {}); err == nil { // not JSON-marshalable
		t.Fatal("expected an error for an unmarshalable sample")
	}
}

func TestValidateSchemaWildcardTarget(t *testing.T) {
	s := New()
	if err := s.LoadBytes([]byte(`{"fields":[{"target":"tags.*"}]}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateSchema(map[string]any{"tags": []any{"a", "b"}}); err != nil {
		t.Fatalf("expected wildcard target to match, got %v", err)
	}
	if err := s.ValidateSchema(map[string]any{"other": 1}); err == nil {
		t.Fatal("expected wildcard target with no match in sample to be flagged")
	}
}
