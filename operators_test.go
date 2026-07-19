package schematics

import "testing"

func TestOperators(t *testing.T) {
	ctx := &Context{}
	if v, _ := opTrim("  hi  ", nil, ctx); v != "hi" {
		t.Errorf("trim: %v", v)
	}
	if v, _ := opCapitalize("hELLO", nil, ctx); v != "Hello" {
		t.Errorf("capitalize: %v", v)
	}
	// capitalize must not panic on empty string (v1 did str[0] unguarded).
	if v, _ := opCapitalize("", nil, ctx); v != "" {
		t.Errorf("capitalize empty: %v", v)
	}
	if v, _ := opUpper("hi", nil, ctx); v != "HI" {
		t.Errorf("upper: %v", v)
	}
	if v, _ := opAdd(2.0, Args{"value": 3.0}, ctx); v != 5.0 {
		t.Errorf("add: %v", v)
	}
	if _, err := opDivide(2.0, Args{"value": 0.0}, ctx); err == nil {
		t.Errorf("divide by zero should error")
	}
	// add on a non-number returns an error rather than panicking.
	if _, err := opAdd("x", Args{"value": 1.0}, ctx); err == nil {
		t.Errorf("add on string should error")
	}
	if v, _ := opDefault("", Args{"value": "fallback"}, ctx); v != "fallback" {
		t.Errorf("default: %v", v)
	}
	if v, _ := opRound(3.14159, Args{"places": 2.0}, ctx); v != 3.14 {
		t.Errorf("round: %v", v)
	}
}
