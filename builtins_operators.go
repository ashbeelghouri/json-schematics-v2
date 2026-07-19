package schematics

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

func opTrim(v any, _ Args, _ *Context) (any, error) {
	s, ok := asString(v)
	if !ok {
		return v, nil
	}
	return strings.TrimSpace(s), nil
}

func opCapitalize(v any, _ Args, _ *Context) (any, error) {
	s, ok := asString(v)
	if !ok {
		return v, nil
	}
	if s == "" {
		return s, nil
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + strings.ToLower(string(r[1:])), nil
}

func opUpper(v any, _ Args, _ *Context) (any, error) {
	s, ok := asString(v)
	if !ok {
		return v, nil
	}
	return strings.ToUpper(s), nil
}

func opLower(v any, _ Args, _ *Context) (any, error) {
	s, ok := asString(v)
	if !ok {
		return v, nil
	}
	return strings.ToLower(s), nil
}

func opToString(v any, _ Args, _ *Context) (any, error) {
	return toString(v), nil
}

func opAdd(v any, args Args, _ *Context) (any, error) {
	n, ok := toFloat(v)
	if !ok {
		return nil, errNotNumber
	}
	by, err := args.Float("value")
	if err != nil {
		return nil, err
	}
	return n + by, nil
}

func opSubtract(v any, args Args, _ *Context) (any, error) {
	n, ok := toFloat(v)
	if !ok {
		return nil, errNotNumber
	}
	by, err := args.Float("value")
	if err != nil {
		return nil, err
	}
	return n - by, nil
}

func opMultiply(v any, args Args, _ *Context) (any, error) {
	n, ok := toFloat(v)
	if !ok {
		return nil, errNotNumber
	}
	by, err := args.Float("value")
	if err != nil {
		return nil, err
	}
	return n * by, nil
}

func opDivide(v any, args Args, _ *Context) (any, error) {
	n, ok := toFloat(v)
	if !ok {
		return nil, errNotNumber
	}
	by, err := args.Float("value")
	if err != nil {
		return nil, err
	}
	if by == 0 {
		return nil, errors.New("division by zero")
	}
	return n / by, nil
}

func opRound(v any, args Args, _ *Context) (any, error) {
	n, ok := toFloat(v)
	if !ok {
		return nil, errNotNumber
	}
	places := args.FloatOr("places", 0)
	factor := math.Pow(10, places)
	return math.Round(n*factor) / factor, nil
}

func opDefault(v any, args Args, _ *Context) (any, error) {
	def, ok := args.Get("value")
	if !ok {
		return nil, errors.New("missing required argument \"value\"")
	}
	if v == nil {
		return def, nil
	}
	if s, ok := v.(string); ok && s == "" {
		return def, nil
	}
	return v, nil
}

// opArrayToObject converts an array of objects into an object keyed by each
// element's "key" field.
func opArrayToObject(v any, args Args, _ *Context) (any, error) {
	arr, ok := asSlice(v)
	if !ok {
		return v, nil
	}
	keyField, err := args.String("key")
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("array element is not an object")
		}
		k, ok := obj[keyField].(string)
		if !ok {
			return nil, fmt.Errorf("array element is missing string key %q", keyField)
		}
		out[k] = obj
	}
	return out, nil
}
