package schematics

import "fmt"

// Args holds the arguments a schema passes to a validator, operator, or
// condition. The typed accessors never panic: a missing key or a wrong type is
// reported as an error instead, which is what makes built-in and custom rules
// safe to run against arbitrary data.
type Args map[string]any

// Has reports whether key is present.
func (a Args) Has(key string) bool {
	_, ok := a[key]
	return ok
}

// String returns the string argument at key.
func (a Args) String(key string) (string, error) {
	v, ok := a[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("argument %q must be a string", key)
	}
	return s, nil
}

// Float returns the numeric argument at key as a float64.
func (a Args) Float(key string) (float64, error) {
	v, ok := a[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	f, ok := toFloat(v)
	if !ok {
		return 0, fmt.Errorf("argument %q must be a number", key)
	}
	return f, nil
}

// Int returns the numeric argument at key truncated to an int.
func (a Args) Int(key string) (int, error) {
	f, err := a.Float(key)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

// Bool returns the boolean argument at key.
func (a Args) Bool(key string) (bool, error) {
	v, ok := a[key]
	if !ok {
		return false, fmt.Errorf("missing required argument %q", key)
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("argument %q must be a boolean", key)
	}
	return b, nil
}

// Strings returns the argument at key as a slice of strings.
func (a Args) Strings(key string) ([]string, error) {
	v, ok := a[key]
	if !ok {
		return nil, fmt.Errorf("missing required argument %q", key)
	}
	switch s := v.(type) {
	case []string:
		return s, nil
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("argument %q must be an array of strings", key)
			}
			out = append(out, str)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("argument %q must be an array of strings", key)
	}
}

// StringOr returns the string argument at key, or def if absent or mistyped.
func (a Args) StringOr(key, def string) string {
	if s, err := a.String(key); err == nil {
		return s
	}
	return def
}

// FloatOr returns the numeric argument at key, or def if absent or mistyped.
func (a Args) FloatOr(key string, def float64) float64 {
	if f, err := a.Float(key); err == nil {
		return f
	}
	return def
}

// Get returns the raw argument at key.
func (a Args) Get(key string) (any, bool) {
	v, ok := a[key]
	return v, ok
}
