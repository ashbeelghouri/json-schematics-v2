package schematics

import (
	"errors"
	"fmt"
	"math"
)

var errNotNumber = errors.New("value is not a number")

func ruleIsNumber(v any, _ Args, _ *Context) error {
	if _, ok := toFloat(v); !ok {
		return errNotNumber
	}
	return nil
}

func ruleIsInteger(v any, _ Args, _ *Context) error {
	f, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	if f != math.Trunc(f) {
		return errors.New("must be an integer")
	}
	return nil
}

func ruleIsFloat(v any, _ Args, _ *Context) error {
	if _, ok := toFloat(v); !ok {
		return errNotNumber
	}
	return nil
}

func ruleMax(v any, args Args, _ *Context) error {
	n, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	max, err := args.Float("max")
	if err != nil {
		return err
	}
	if n > max {
		return fmt.Errorf("must be at most %v", max)
	}
	return nil
}

func ruleMin(v any, args Args, _ *Context) error {
	n, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	min, err := args.Float("min")
	if err != nil {
		return err
	}
	if n < min {
		return fmt.Errorf("must be at least %v", min)
	}
	return nil
}

func ruleBetween(v any, args Args, ctx *Context) error {
	if err := ruleMin(v, args, ctx); err != nil {
		return err
	}
	return ruleMax(v, args, ctx)
}

func rulePositive(v any, _ Args, _ *Context) error {
	n, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	if n <= 0 {
		return errors.New("must be greater than zero")
	}
	return nil
}

func ruleNegative(v any, _ Args, _ *Context) error {
	n, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	if n >= 0 {
		return errors.New("must be less than zero")
	}
	return nil
}

func ruleNonNegative(v any, _ Args, _ *Context) error {
	n, ok := toFloat(v)
	if !ok {
		return errNotNumber
	}
	if n < 0 {
		return errors.New("must not be negative")
	}
	return nil
}
