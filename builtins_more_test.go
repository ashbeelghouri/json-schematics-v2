package schematics

import "testing"

func TestMoreStringRules(t *testing.T) {
	c := &Context{}
	ok(t, ruleNotEmpty("x", nil, c), "notEmpty")
	bad(t, ruleNotEmpty("  ", nil, c), "notEmpty blank")
	ok(t, ruleLengthBetween("abcd", Args{"min": 2.0, "max": 5.0}, c), "lengthBetween")
	bad(t, ruleLengthBetween("a", Args{"min": 2.0, "max": 5.0}, c), "lengthBetween under")
	ok(t, ruleHasUpper("aA", nil, c), "hasUpper")
	ok(t, ruleHasLower("Aa", nil, c), "hasLower")
	ok(t, ruleHasDigit("a1", nil, c), "hasDigit")
	ok(t, ruleNotURL("hello", nil, c), "notURL")
	bad(t, ruleNotURL("https://x.com", nil, c), "notURL on url")
	ok(t, ruleIsHTTPS("https://x.com", nil, c), "isHTTPS")
	bad(t, ruleIsHTTPS("http://x.com", nil, c), "isHTTPS on http")
	ok(t, ruleURLHasHost("https://api.x.com/y", Args{"host": "x.com"}, c), "urlHasHost")
	bad(t, ruleURLHasHost("https://api.z.com", Args{"host": "x.com"}, c), "urlHasHost mismatch")
	ok(t, ruleURLHasQuery("https://x.com?a=1&b=2", Args{"params": "a,b"}, c), "urlHasQuery")
	bad(t, ruleURLHasQuery("https://x.com?a=1", Args{"params": []any{"a", "b"}}, c), "urlHasQuery missing")
	ok(t, ruleEquals("yes", Args{"value": "yes"}, c), "equals")
	bad(t, ruleEquals("no", Args{"value": "yes"}, c), "equals mismatch")
	ok(t, ruleInOptions("b", Args{"options": []any{"a", "b"}}, c), "inOptions")
	ok(t, ruleMatchRegex("abc123", Args{"pattern": `^[a-z]+\d+$`}, c), "matchRegex")
	bad(t, ruleMatchRegex("ABC", Args{"pattern": `^[a-z]+$`}, c), "matchRegex fail")
	ok(t, ruleLike("hello world", Args{"pattern": "hello%"}, c), "like %")
	ok(t, ruleLike("cat", Args{"pattern": "c_t"}, c), "like _")
	bad(t, ruleLike("dog", Args{"pattern": "c_t"}, c), "like fail")
}

func TestMoreNumberAndDate(t *testing.T) {
	c := &Context{}
	ok(t, ruleIsInteger(5.0, nil, c), "isInteger")
	bad(t, ruleIsInteger(5.5, nil, c), "isInteger frac")
	ok(t, ruleBetween(5.0, Args{"min": 1.0, "max": 10.0}, c), "between")
	bad(t, ruleBetween(50.0, Args{"min": 1.0, "max": 10.0}, c), "between over")
	ok(t, ruleNonNegative(0.0, nil, c), "nonNegative")
	ok(t, ruleBetweenTime("2020-06-01", Args{"min": "2020-01-01", "max": "2020-12-31"}, c), "betweenTime")
	bad(t, ruleBetweenTime("2021-06-01", Args{"min": "2020-01-01", "max": "2020-12-31"}, c), "betweenTime out")
}

func TestMoreOperatorsAndConditions(t *testing.T) {
	c := &Context{Separator: ".", Flat: map[string]any{"a": "1", "b": "x"}}
	if v, _ := opSubtract(5.0, Args{"value": 2.0}, c); v != 3.0 {
		t.Errorf("subtract %v", v)
	}
	if v, _ := opMultiply(4.0, Args{"value": 2.0}, c); v != 8.0 {
		t.Errorf("multiply %v", v)
	}
	if v, _ := opToString(42.0, nil, c); v != "42" {
		t.Errorf("toString %v", v)
	}
	obj, err := opArrayToObject([]any{
		map[string]any{"k": "one", "v": 1.0},
		map[string]any{"k": "two", "v": 2.0},
	}, Args{"key": "k"}, c)
	if err != nil {
		t.Fatalf("arrayToObject: %v", err)
	}
	m := obj.(map[string]any)
	if _, ok := m["one"]; !ok {
		t.Errorf("arrayToObject missing key: %v", m)
	}

	if !condFieldPresent(Args{"field": "a"}, c) {
		t.Error("fieldPresent should be true")
	}
	if condFieldAbsent(Args{"field": "a"}, c) {
		t.Error("fieldAbsent should be false")
	}
	if !condFieldEquals(Args{"field": "a", "value": "1"}, c) {
		t.Error("fieldEquals should be true")
	}
	if condFieldEquals(Args{"field": "b", "value": "nope"}, c) {
		t.Error("fieldEquals should be false")
	}
}

func TestArgsAccessors(t *testing.T) {
	a := Args{"s": "x", "n": 3.0, "b": true, "arr": []any{"p", "q"}}
	if _, err := a.String("s"); err != nil {
		t.Error(err)
	}
	if _, err := a.String("missing"); err == nil {
		t.Error("expected missing error")
	}
	if _, err := a.String("n"); err == nil {
		t.Error("expected type error")
	}
	if v, _ := a.Int("n"); v != 3 {
		t.Errorf("Int %d", v)
	}
	if v, _ := a.Bool("b"); !v {
		t.Error("Bool")
	}
	if v, _ := a.Strings("arr"); len(v) != 2 {
		t.Errorf("Strings %v", v)
	}
	if a.StringOr("missing", "def") != "def" {
		t.Error("StringOr default")
	}
}
