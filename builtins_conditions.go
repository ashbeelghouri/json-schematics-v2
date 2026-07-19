package schematics

import "fmt"

// condFieldPresent passes when the target named by the "field" argument is
// present in the document.
func condFieldPresent(args Args, ctx *Context) bool {
	target, err := args.String("field")
	if err != nil {
		return false
	}
	return ctx.FieldPresent(target)
}

// condFieldAbsent is the inverse of condFieldPresent.
func condFieldAbsent(args Args, ctx *Context) bool {
	target, err := args.String("field")
	if err != nil {
		return false
	}
	return !ctx.FieldPresent(target)
}

// condFieldEquals passes when the target named by "field" holds "value".
func condFieldEquals(args Args, ctx *Context) bool {
	target, err := args.String("field")
	if err != nil {
		return false
	}
	want, ok := args.Get("value")
	if !ok {
		return false
	}
	got, ok := ctx.Lookup(target)
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", got) == fmt.Sprintf("%v", want)
}
