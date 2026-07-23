package schematics

import (
	"errors"
	"testing"
)

func TestImportSchema(t *testing.T) {
	s, err := ImportSchema([]byte(`{"fields":[
		{"target":"name","required":true},
		{"target":"email","validate":[{"rule":"email"}]}]}`))
	if err != nil {
		t.Fatalf("ImportSchema: %v", err)
	}
	if err := s.Validate(map[string]any{"name": "Ada", "email": "ada@example.com"}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestImportSchemaBadBytes(t *testing.T) {
	if _, err := ImportSchema([]byte(`not json`)); err == nil {
		t.Fatal("expected an error for malformed schema bytes")
	}
}

func TestImportSchemaAppliesOptions(t *testing.T) {
	s, err := ImportSchema([]byte(`{"fields":[{"target":"n"}]}`), WithLocale("ar"))
	if err != nil {
		t.Fatalf("ImportSchema: %v", err)
	}
	if s.locale != "ar" {
		t.Errorf("expected locale ar, got %q", s.locale)
	}
}

func TestValidateBytesObject(t *testing.T) {
	s, err := ImportSchema([]byte(`{"fields":[{"target":"name","required":true}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateBytes([]byte(`{"name":"Ada"}`), false); err != nil {
		t.Fatalf("expected valid object, got %v", err)
	}
	err = s.ValidateBytes([]byte(`{}`), false)
	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestValidateBytesArray(t *testing.T) {
	s, err := ImportSchema([]byte(`{"arrayIdKey":"id","fields":[{"target":"email","required":true,"validate":[{"rule":"email"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateBytes([]byte(`[{"id":"a","email":"a@x.com"},{"id":"b","email":"b@x.com"}]`), true); err != nil {
		t.Fatalf("expected valid array, got %v", err)
	}
	err = s.ValidateBytes([]byte(`[{"id":"a","email":"a@x.com"},{"id":"b","email":"bad"}]`), true)
	var ve *ValidationErrors
	if !errors.As(err, &ve) || ve.Len() != 1 {
		t.Fatalf("expected 1 error, got %v", err)
	}
	if ve.Errors[0].RowID != "b" {
		t.Errorf("expected RowID b, got %q", ve.Errors[0].RowID)
	}
}

func TestValidateBytesWrongShape(t *testing.T) {
	s, err := ImportSchema([]byte(`{"fields":[{"target":"name"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	// asked for an object but got an array
	if err := s.ValidateBytes([]byte(`[{"name":"Ada"}]`), false); err == nil {
		t.Fatal("expected a parse error for array bytes parsed as an object")
	}
	// asked for an array but got an object
	if err := s.ValidateBytes([]byte(`{"name":"Ada"}`), true); err == nil {
		t.Fatal("expected a parse error for object bytes parsed as an array")
	}
}

func TestValidateBytesMalformedJSON(t *testing.T) {
	s, err := ImportSchema([]byte(`{"fields":[{"target":"name"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ValidateBytes([]byte(`not json`), false); err == nil {
		t.Fatal("expected an error for malformed data bytes")
	}
}
