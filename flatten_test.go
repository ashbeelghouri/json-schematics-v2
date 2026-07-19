package schematics

import (
	"reflect"
	"testing"
)

func TestFlattenAndDeflate(t *testing.T) {
	in := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{"name": "ash", "age": 30.0},
			"tags":    []any{"a", "b"},
		},
	}
	flat := Flatten(in, ".")
	want := map[string]any{
		"user.profile.name": "ash",
		"user.profile.age":  30.0,
		"user.tags.0":       "a",
		"user.tags.1":       "b",
	}
	if !reflect.DeepEqual(flat, want) {
		t.Fatalf("flatten mismatch:\n got %#v\nwant %#v", flat, want)
	}
	back := Deflate(flat, ".")
	if !reflect.DeepEqual(back, in) {
		t.Fatalf("deflate did not round-trip:\n got %#v\nwant %#v", back, in)
	}
}

func TestMatchTargetWildcardAndSubtree(t *testing.T) {
	flat := map[string]any{
		"users.0.name": "a",
		"users.1.name": "b",
		"tags.0":       "x",
		"tags.1":       "y",
	}
	// wildcard on each element
	m := matchTarget(flat, "users.*.name", ".", false)
	if len(m) != 2 || m["users.0.name"] != "a" || m["users.1.name"] != "b" {
		t.Fatalf("wildcard match wrong: %#v", m)
	}
	// subtree reconstruction into a slice for array validators
	m = matchTarget(flat, "tags", ".", false)
	arr, ok := m["tags"].([]any)
	if !ok || len(arr) != 2 || arr[0] != "x" || arr[1] != "y" {
		t.Fatalf("subtree array match wrong: %#v", m)
	}
}

func TestMatchTargetRegex(t *testing.T) {
	flat := map[string]any{"a1": 1.0, "a2": 2.0, "b1": 3.0}
	m := matchTarget(flat, `^a\d$`, ".", true)
	if len(m) != 2 {
		t.Fatalf("regex match wrong: %#v", m)
	}
}
