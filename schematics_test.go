package schematics

import (
	"errors"
	"testing"
)

const personSchema = `{
  "version": "2.0",
  "separator": ".",
  "fields": [
    { "target": "user.profile.name", "required": true,
      "validate": [ {"rule":"isString"}, {"rule":"minLength","args":{"min":2}} ],
      "operate":  [ {"op":"trim"}, {"op":"capitalize"} ] },
    { "target": "user.profile.age", "required": true,
      "validate": [ {"rule":"isNumber"}, {"rule":"max","args":{"max":120},"message":"age too high"} ] },
    { "target": "user.profile.email",
      "dependsOn": ["user.profile.name"],
      "validate": [ {"rule":"email"} ] }
  ]
}`

func newPerson(t *testing.T) *Schematics {
	t.Helper()
	s := New()
	if err := s.LoadBytes([]byte(personSchema)); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := s.Check(); err != nil {
		t.Fatalf("check: %v", err)
	}
	return s
}

func TestValidateHappyPath(t *testing.T) {
	s := newPerson(t)
	data := map[string]any{"user": map[string]any{"profile": map[string]any{
		"name": "ash", "age": 30, "email": "a@b.com",
	}}}
	if err := s.Validate(data); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateRequiredSurfaces(t *testing.T) {
	// v1 bug: required-field failures were silently dropped.
	s := newPerson(t)
	data := map[string]any{"user": map[string]any{"profile": map[string]any{
		"age": 30,
	}}}
	err := s.Validate(data)
	if err == nil {
		t.Fatal("expected required error")
	}
	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}
	found := false
	for _, e := range ve.Errors {
		if e.Target == "user.profile.name" && e.Rule == "required" {
			found = true
		}
	}
	if !found {
		t.Fatalf("required error for name not surfaced: %v", ve.Strings("en", "%target %rule"))
	}
}

func TestCustomMessageAndLocale(t *testing.T) {
	s := New(WithLocale("en"))
	_ = s.LoadBytes([]byte(`{"fields":[
		{"target":"n","validate":[{"rule":"minLength","args":{"min":5},
		 "message":"too short","messages":{"ar":"قصير جدا"}}]}]}`))
	err := s.Validate(map[string]any{"n": "hi"})
	var ve *ValidationErrors
	if !errors.As(err, &ve) || ve.Len() != 1 {
		t.Fatalf("expected 1 error, got %v", err)
	}
	if got := ve.Errors[0].Message("en"); got != "too short" {
		t.Errorf("en message: %q", got)
	}
	if got := ve.Errors[0].Message("ar"); got != "قصير جدا" {
		t.Errorf("ar message: %q", got)
	}
}

func TestDependsOn(t *testing.T) {
	s := newPerson(t)
	// email present but its dependency (name) absent -> dependency error.
	data := map[string]any{"user": map[string]any{"profile": map[string]any{
		"age": 30, "email": "a@b.com",
	}}}
	err := s.Validate(data)
	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected errors, got %v", err)
	}
	hasDep := false
	for _, e := range ve.Errors {
		if e.Target == "user.profile.email" && e.Rule == "dependsOn" {
			hasDep = true
		}
	}
	if !hasDep {
		t.Errorf("expected dependsOn error: %v", ve.Strings("en", "%target:%rule"))
	}
}

func TestConditionsGate(t *testing.T) {
	// last name only validated when first name present.
	schema := `{"fields":[
		{"target":"last",
		 "when":[{"condition":"fieldPresent","args":{"field":"first"}}],
		 "validate":[{"rule":"minLength","args":{"min":3}}]}]}`
	s := New()
	if err := s.LoadBytes([]byte(schema)); err != nil {
		t.Fatal(err)
	}
	// first absent -> condition false -> last not validated even though short.
	if err := s.Validate(map[string]any{"last": "xy"}); err != nil {
		t.Fatalf("condition should have skipped validation: %v", err)
	}
	// first present -> last validated -> fails.
	if err := s.Validate(map[string]any{"first": "a", "last": "xy"}); err == nil {
		t.Fatal("expected last to be validated and fail")
	}
}

func TestAddToDBNoPanic(t *testing.T) {
	// v1 bug: add_to_db wrote into a nil DB map and panicked.
	schema := `{"fields":[
		{"target":"threshold","addToDB":true},
		{"target":"value","validate":[{"rule":"isNumber"}]}]}`
	s := New()
	if err := s.LoadBytes([]byte(schema)); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(map[string]any{"threshold": 10, "value": 5}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestOperateEndToEnd(t *testing.T) {
	s := newPerson(t)
	data := map[string]any{"user": map[string]any{"profile": map[string]any{
		"name": "  aSHBEEL  ",
	}}}
	out, err := s.Operate(data)
	if err != nil {
		t.Fatalf("operate: %v", err)
	}
	m := out.(map[string]any)
	got := m["user"].(map[string]any)["profile"].(map[string]any)["name"]
	if got != "Ashbeel" {
		t.Errorf("operate chain trim+capitalize: %q", got)
	}
}

func TestArrayValidationWithRowID(t *testing.T) {
	schema := `{"version":"2.0","arrayIdKey":"id","fields":[
		{"target":"email","required":true,"validate":[{"rule":"email"}]}]}`
	s := New()
	if err := s.LoadBytes([]byte(schema)); err != nil {
		t.Fatal(err)
	}
	data := []map[string]any{
		{"id": "row-a", "email": "good@x.com"},
		{"id": "row-b", "email": "bad"},
	}
	err := s.Validate(data)
	var ve *ValidationErrors
	if !errors.As(err, &ve) || ve.Len() != 1 {
		t.Fatalf("expected 1 error, got %v", err)
	}
	if ve.Errors[0].RowID != "row-b" {
		t.Errorf("expected RowID row-b, got %q", ve.Errors[0].RowID)
	}
}

func TestArrayItemsValidator(t *testing.T) {
	// subtree reconstruction feeds an array validator.
	schema := `{"fields":[
		{"target":"tags","validate":[{"rule":"minItems","args":{"min":2}},{"rule":"unique"}]}]}`
	s := New()
	if err := s.LoadBytes([]byte(schema)); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(map[string]any{"tags": []any{"a", "b"}}); err != nil {
		t.Fatalf("valid tags rejected: %v", err)
	}
	if err := s.Validate(map[string]any{"tags": []any{"a"}}); err == nil {
		t.Fatal("expected minItems failure")
	}
}

func TestUnknownRuleIsSchemaError(t *testing.T) {
	s := New()
	_ = s.LoadBytes([]byte(`{"fields":[{"target":"x","validate":[{"rule":"noSuchRule"}]}]}`))
	err := s.Validate(map[string]any{"x": "y"})
	var se *SchemaError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SchemaError, got %T: %v", err, err)
	}
}

func TestCustomRule(t *testing.T) {
	s := New()
	s.RegisterRule("isAsh", func(v any, _ Args, _ *Context) error {
		if v == "ash" {
			return nil
		}
		return errors.New("not ash")
	})
	_ = s.LoadBytes([]byte(`{"fields":[{"target":"n","validate":[{"rule":"isAsh"}]}]}`))
	if err := s.Validate(map[string]any{"n": "ash"}); err != nil {
		t.Fatalf("custom rule pass: %v", err)
	}
	if err := s.Validate(map[string]any{"n": "bob"}); err == nil {
		t.Fatal("custom rule should fail")
	}
}
