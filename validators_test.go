package schematics

import "testing"

func ok(t *testing.T, err error, name string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: expected pass, got %v", name, err)
	}
}
func bad(t *testing.T, err error, name string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected failure, got nil", name)
	}
}

func TestStringValidators(t *testing.T) {
	ctx := &Context{}
	ok(t, ruleIsString("hi", nil, ctx), "isString")
	bad(t, ruleIsString(5, nil, ctx), "isString non-string")
	ok(t, ruleEmail("a@b.com", nil, ctx), "email")
	bad(t, ruleEmail("nope", nil, ctx), "email bad")
	ok(t, ruleMaxLength("héllo", Args{"max": 5.0}, ctx), "maxLength rune count")
	bad(t, ruleMaxLength("toolong", Args{"max": 3.0}, ctx), "maxLength over")

	// v1 bug: NoSpecialChars was inverted. Correct behavior:
	ok(t, ruleNoSpecialChars("abc123", nil, ctx), "noSpecialChars clean")
	bad(t, ruleNoSpecialChars("abc!", nil, ctx), "noSpecialChars dirty")
	ok(t, ruleHasSpecialChars("a!b", nil, ctx), "hasSpecialChars")
	bad(t, ruleHasSpecialChars("abc", nil, ctx), "hasSpecialChars none")

	// v1 bug: IsURL name was overwritten by the UUID validator.
	ok(t, ruleIsURL("https://x.com/y", nil, ctx), "isURL")
	bad(t, ruleIsURL("nope", nil, ctx), "isURL bad")
	ok(t, ruleIsUUID("123e4567-e89b-12d3-a456-426614174000", nil, ctx), "isUUID")
	bad(t, ruleIsUUID("not-a-uuid", nil, ctx), "isUUID bad")

	// non-string inputs must return errors, never panic.
	bad(t, ruleMaxLength(5, Args{"max": 3.0}, ctx), "maxLength non-string no panic")
}

func TestNumberValidators(t *testing.T) {
	ctx := &Context{}
	ok(t, ruleIsNumber(3.0, nil, ctx), "isNumber")
	bad(t, ruleIsNumber("x", nil, ctx), "isNumber bad")
	ok(t, ruleMax(3.0, Args{"max": 5.0}, ctx), "max")
	bad(t, ruleMax(9.0, Args{"max": 5.0}, ctx), "max over")

	// v1 bug: IsLesserThanZero passed {"max":0} to a min-only function.
	ok(t, ruleNegative(-2.0, nil, ctx), "negative")
	bad(t, ruleNegative(2.0, nil, ctx), "negative on positive")
	ok(t, rulePositive(2.0, nil, ctx), "positive")
	bad(t, rulePositive(0.0, nil, ctx), "positive on zero")
}

func TestDateValidators(t *testing.T) {
	ctx := &Context{}
	ok(t, ruleIsDate("2020-01-02", nil, ctx), "isDate")
	bad(t, ruleIsDate("nope", nil, ctx), "isDate bad")
	ok(t, ruleBeforeNow("2000-01-01", nil, ctx), "beforeNow")
	bad(t, ruleAfterNow("2000-01-01", nil, ctx), "afterNow past")

	// v1 bug: Before/After both read "maxTime" and both said "is after".
	ok(t, ruleBefore("2020-01-01", Args{"date": "2020-06-01"}, ctx), "before")
	bad(t, ruleBefore("2020-12-01", Args{"date": "2020-06-01"}, ctx), "before fail")
	ok(t, ruleAfter("2020-12-01", Args{"date": "2020-06-01"}, ctx), "after")
	bad(t, ruleAfter("2020-01-01", Args{"date": "2020-06-01"}, ctx), "after fail")
}

func TestArrayValidators(t *testing.T) {
	ctx := &Context{}
	ok(t, ruleIsArray([]any{1, 2}, nil, ctx), "isArray")
	bad(t, ruleIsArray("x", nil, ctx), "isArray bad")
	ok(t, ruleMaxItems([]any{1, 2}, Args{"max": 3.0}, ctx), "maxItems")
	bad(t, ruleMinItems([]any{1}, Args{"min": 2.0}, ctx), "minItems under")
	ok(t, ruleUnique([]any{"a", "b"}, nil, ctx), "unique")
	bad(t, ruleUnique([]any{"a", "a"}, nil, ctx), "unique dup")
	ok(t, ruleItemsInOptions([]any{"a"}, Args{"options": []any{"a", "b"}}, ctx), "itemsInOptions")
	bad(t, ruleItemsInOptions([]any{"z"}, Args{"options": []any{"a", "b"}}, ctx), "itemsInOptions bad")
}
