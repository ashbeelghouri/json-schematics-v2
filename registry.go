package schematics

// Rule validates a value. It returns nil when the value is acceptable, or an
// error describing why it is not. Rules must never panic: use the typed Args
// accessors and the value-inspection helpers, both of which report bad input as
// errors.
type Rule func(value any, args Args, ctx *Context) error

// Operator transforms a value and returns the replacement. Returning an error
// aborts the whole Operate call.
type Operator func(value any, args Args, ctx *Context) (any, error)

// Condition decides whether a field's validators and operators should run.
type Condition func(args Args, ctx *Context) bool

// RegisterRule adds or replaces a named validator. It is safe to call after
// loading a schema and before Validate.
func (s *Schematics) RegisterRule(name string, fn Rule) *Schematics {
	s.rules[name] = fn
	s.checked = false
	return s
}

// RegisterOperator adds or replaces a named operator.
func (s *Schematics) RegisterOperator(name string, fn Operator) *Schematics {
	s.operators[name] = fn
	s.checked = false
	return s
}

// RegisterCondition adds or replaces a named condition.
func (s *Schematics) RegisterCondition(name string, fn Condition) *Schematics {
	s.conditions[name] = fn
	s.checked = false
	return s
}

// RuleNames returns the names of every registered validator.
func (s *Schematics) RuleNames() []string { return keysOf(s.rules) }

// OperatorNames returns the names of every registered operator.
func (s *Schematics) OperatorNames() []string { return keysOf(s.operators) }

// ConditionNames returns the names of every registered condition.
func (s *Schematics) ConditionNames() []string { return keysOf(s.conditions) }

func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
