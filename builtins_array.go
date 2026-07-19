package schematics

import (
	"errors"
	"fmt"
)

var errNotArray = errors.New("value is not an array")

func ruleIsArray(v any, _ Args, _ *Context) error {
	if _, ok := asSlice(v); !ok {
		return errNotArray
	}
	return nil
}

func ruleMaxItems(v any, args Args, _ *Context) error {
	arr, ok := asSlice(v)
	if !ok {
		return errNotArray
	}
	max, err := args.Int("max")
	if err != nil {
		return err
	}
	if len(arr) > max {
		return fmt.Errorf("must have at most %d items", max)
	}
	return nil
}

func ruleMinItems(v any, args Args, _ *Context) error {
	arr, ok := asSlice(v)
	if !ok {
		return errNotArray
	}
	min, err := args.Int("min")
	if err != nil {
		return err
	}
	if len(arr) < min {
		return fmt.Errorf("must have at least %d items", min)
	}
	return nil
}

func ruleUnique(v any, _ Args, _ *Context) error {
	arr, ok := asSlice(v)
	if !ok {
		return errNotArray
	}
	seen := make(map[string]bool, len(arr))
	for _, item := range arr {
		key := fmt.Sprintf("%v", item)
		if seen[key] {
			return fmt.Errorf("items must be unique (duplicate %v)", item)
		}
		seen[key] = true
	}
	return nil
}

func ruleItemsInOptions(v any, args Args, _ *Context) error {
	arr, ok := asSlice(v)
	if !ok {
		return errNotArray
	}
	opts, err := args.Strings("options")
	if err != nil {
		return err
	}
	allowed := make(map[string]bool, len(opts))
	for _, o := range opts {
		allowed[o] = true
	}
	for _, item := range arr {
		s, ok := asString(item)
		if !ok {
			return fmt.Errorf("array item %v is not a string", item)
		}
		if !allowed[s] {
			return fmt.Errorf("%q is not one of the allowed values", s)
		}
	}
	return nil
}
