package schematics

import "context"

// Context is handed to every validator, operator, and condition. It carries the
// shared DB, the active locale and separator, the full flattened document, and
// metadata about the field currently being processed. Ctx is a standard
// context.Context so long-running custom rules can honor cancellation.
type Context struct {
	Ctx       context.Context
	DB        map[string]any
	Locale    string
	Separator string
	Flat      map[string]any
	RowID     string
	Field     *FieldView
}

// FieldView is a read-only snapshot of the field a rule is running against.
type FieldView struct {
	Target   string
	Name     string
	Type     string
	Required bool
	Provided bool
	Tags     []string
}

// FieldPresent reports whether target selects at least one value in the
// document under validation. It honors the active separator and wildcard rules.
func (c *Context) FieldPresent(target string) bool {
	return len(matchTarget(c.Flat, target, c.Separator, false)) > 0
}

// Lookup returns the first value selected by target, if any.
func (c *Context) Lookup(target string) (any, bool) {
	m := matchTarget(c.Flat, target, c.Separator, false)
	for _, v := range m {
		return v, true
	}
	return nil, false
}
