package schematics

import (
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Flatten converts a nested map into a single-level map whose keys are the paths
// to each leaf value, joined by separator. Arrays become indexed keys, so
// {"tags":["a","b"]} becomes {"tags.0":"a","tags.1":"b"}. Empty maps and slices
// are preserved as leaf values so length-style validators can still see them.
func Flatten(data map[string]any, separator string) map[string]any {
	if separator == "" {
		separator = defaultSeparator
	}
	out := make(map[string]any)
	flattenInto(out, "", data, separator)
	return out
}

func flattenInto(out map[string]any, prefix string, value any, sep string) {
	if value == nil {
		if prefix != "" {
			out[prefix] = nil
		}
		return
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		if rv.Len() == 0 {
			if prefix != "" {
				out[prefix] = value
			}
			return
		}
		for _, mk := range rv.MapKeys() {
			key := toString(mk.Interface())
			if prefix != "" {
				key = prefix + sep + key
			}
			flattenInto(out, key, rv.MapIndex(mk).Interface(), sep)
		}
	case reflect.Slice, reflect.Array:
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
			// treat []byte as an opaque leaf
			if prefix != "" {
				out[prefix] = value
			}
			return
		}
		if rv.Len() == 0 {
			if prefix != "" {
				out[prefix] = value
			}
			return
		}
		for i := 0; i < rv.Len(); i++ {
			key := strconv.Itoa(i)
			if prefix != "" {
				key = prefix + sep + key
			}
			flattenInto(out, key, rv.Index(i).Interface(), sep)
		}
	default:
		if prefix != "" {
			out[prefix] = value
		}
	}
}

// Deflate reconstructs a nested structure from a flattened map. Segments that
// are consecutive integers starting at zero are rebuilt as slices.
func Deflate(flat map[string]any, separator string) map[string]any {
	if separator == "" {
		separator = defaultSeparator
	}
	raw := deflateRaw(flat, separator)
	if res, ok := arrayify(raw).(map[string]any); ok {
		return res
	}
	return raw
}

func deflateRaw(flat map[string]any, sep string) map[string]any {
	root := map[string]any{}
	for key, val := range flat {
		parts := strings.Split(key, sep)
		cur := root
		for i := 0; i < len(parts)-1; i++ {
			p := parts[i]
			next, ok := cur[p].(map[string]any)
			if !ok {
				next = map[string]any{}
				cur[p] = next
			}
			cur = next
		}
		cur[parts[len(parts)-1]] = val
	}
	return root
}

// arrayify recursively converts maps whose keys are exactly {0..n-1} into slices.
func arrayify(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	for k, val := range m {
		m[k] = arrayify(val)
	}
	if len(m) == 0 {
		return m
	}
	idxs := make([]int, 0, len(m))
	for k := range m {
		n, err := strconv.Atoi(k)
		if err != nil || n < 0 {
			return m
		}
		idxs = append(idxs, n)
	}
	sort.Ints(idxs)
	for i, n := range idxs {
		if i != n {
			return m
		}
	}
	arr := make([]any, len(idxs))
	for i := range idxs {
		arr[i] = m[strconv.Itoa(i)]
	}
	return arr
}

// matchTarget returns every flattened key/value that a field target selects.
//
// A target may be a literal dotted path, a path containing a "*" wildcard
// (matching a single segment), or — when asRegex is true — a full regular
// expression matched against the flattened keys. When no key matches directly
// but keys exist beneath the target (target+separator+...), the subtree is
// reconstructed and returned as a single value keyed by the target, so array
// and object validators receive the whole structure.
func matchTarget(flat map[string]any, target, separator string, asRegex bool) map[string]any {
	matches := map[string]any{}
	var re *regexp.Regexp
	if asRegex {
		re, _ = regexp.Compile(target)
	} else {
		re = wildcardRegex(target, separator)
	}
	prefix := target + separator
	subtree := map[string]any{}
	for key, val := range flat {
		switch {
		case re != nil && re.MatchString(key):
			matches[key] = val
		case strings.HasPrefix(key, prefix):
			subtree[strings.TrimPrefix(key, prefix)] = val
		}
	}
	if len(matches) == 0 && len(subtree) > 0 {
		matches[target] = collapseSubtree(subtree, separator)
	}
	return matches
}

func wildcardRegex(target, sep string) *regexp.Regexp {
	quoted := regexp.QuoteMeta(target)
	seg := "[^" + regexp.QuoteMeta(sep) + "]+"
	pattern := "^" + strings.ReplaceAll(quoted, `\*`, seg) + "$"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

func collapseSubtree(subtree map[string]any, sep string) any {
	return arrayify(deflateRaw(subtree, sep))
}
