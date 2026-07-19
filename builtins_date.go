package schematics

import (
	"errors"
	"fmt"
	"time"
)

const dateFmt = "2006-01-02 15:04:05"

var errBadDate = errors.New("invalid date")

func ruleIsDate(v any, _ Args, _ *Context) error {
	if _, ok := parseDate(v); !ok {
		return errBadDate
	}
	return nil
}

func ruleBeforeNow(v any, _ Args, _ *Context) error {
	d, ok := parseDate(v)
	if !ok {
		return errBadDate
	}
	if !d.Before(time.Now()) {
		return fmt.Errorf("%s is not in the past", d.Format(dateFmt))
	}
	return nil
}

func ruleAfterNow(v any, _ Args, _ *Context) error {
	d, ok := parseDate(v)
	if !ok {
		return errBadDate
	}
	if !d.After(time.Now()) {
		return fmt.Errorf("%s is not in the future", d.Format(dateFmt))
	}
	return nil
}

func ruleBefore(v any, args Args, _ *Context) error {
	d, ok := parseDate(v)
	if !ok {
		return errBadDate
	}
	other, err := args.String("date")
	if err != nil {
		return err
	}
	ot, ok := parseDate(other)
	if !ok {
		return fmt.Errorf("argument %q is not a valid date", "date")
	}
	if !d.Before(ot) {
		return fmt.Errorf("%s must be before %s", d.Format(dateFmt), ot.Format(dateFmt))
	}
	return nil
}

func ruleAfter(v any, args Args, _ *Context) error {
	d, ok := parseDate(v)
	if !ok {
		return errBadDate
	}
	other, err := args.String("date")
	if err != nil {
		return err
	}
	ot, ok := parseDate(other)
	if !ok {
		return fmt.Errorf("argument %q is not a valid date", "date")
	}
	if !d.After(ot) {
		return fmt.Errorf("%s must be after %s", d.Format(dateFmt), ot.Format(dateFmt))
	}
	return nil
}

func ruleBetweenTime(v any, args Args, _ *Context) error {
	d, ok := parseDate(v)
	if !ok {
		return errBadDate
	}
	minS, err := args.String("min")
	if err != nil {
		return err
	}
	maxS, err := args.String("max")
	if err != nil {
		return err
	}
	minT, ok1 := parseDate(minS)
	maxT, ok2 := parseDate(maxS)
	if !ok1 || !ok2 {
		return errors.New("min and max must be valid dates")
	}
	if d.Before(minT) || d.After(maxT) {
		return fmt.Errorf("%s must be between %s and %s", d.Format(dateFmt), minT.Format(dateFmt), maxT.Format(dateFmt))
	}
	return nil
}
