package schematics

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidationError describes a single rule that failed on a single target value.
type ValidationError struct {
	// Target is the concrete flattened key that failed, e.g. "user.profile.name".
	Target string
	// Rule is the name of the validator (or "required"/"dependsOn") that failed.
	Rule string
	// Value is the offending value.
	Value any
	// RowID identifies the array row for array inputs; empty for plain objects.
	RowID string

	fallback string            // default-locale message
	messages map[string]string // locale -> message
}

func newError(target, rule string, value any) *ValidationError {
	return &ValidationError{Target: target, Rule: rule, Value: value, messages: map[string]string{}}
}

func (e *ValidationError) setMessage(locale, msg string) {
	if e.messages == nil {
		e.messages = map[string]string{}
	}
	e.messages[locale] = msg
}

// Message returns the message for locale, falling back to the default message
// and then to a generated description.
func (e *ValidationError) Message(locale string) string {
	if m, ok := e.messages[locale]; ok && m != "" {
		return m
	}
	if e.fallback != "" {
		return e.fallback
	}
	if m, ok := e.messages[defaultLocale]; ok && m != "" {
		return m
	}
	return fmt.Sprintf("%s failed for %s", e.Rule, e.Target)
}

// Error implements the error interface using the default locale.
func (e *ValidationError) Error() string {
	return e.Format(defaultLocale, "")
}

// Format renders the error using a template. Recognized tokens are %message,
// %target, %rule (alias %validator), %value and %id.
func (e *ValidationError) Format(locale, format string) string {
	if format == "" {
		format = "%target: %message"
	}
	r := strings.NewReplacer(
		"%message", e.Message(locale),
		"%target", e.Target,
		"%validator", e.Rule,
		"%rule", e.Rule,
		"%value", fmt.Sprintf("%v", e.Value),
		"%id", e.RowID,
	)
	return r.Replace(format)
}

// MarshalJSON renders the error as a stable JSON object.
func (e *ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"target":   e.Target,
		"rule":     e.Rule,
		"value":    e.Value,
		"id":       e.RowID,
		"message":  e.Message(defaultLocale),
		"messages": e.messages,
	})
}

// ValidationErrors is the aggregate error returned by Validate. It implements
// the error interface, so callers can use errors.As to recover it.
type ValidationErrors struct {
	Errors []*ValidationError
}

// Add appends a non-nil error.
func (es *ValidationErrors) Add(e *ValidationError) {
	if e != nil {
		es.Errors = append(es.Errors, e)
	}
}

// HasErrors reports whether any errors were collected.
func (es *ValidationErrors) HasErrors() bool {
	return es != nil && len(es.Errors) > 0
}

// Len returns the number of collected errors.
func (es *ValidationErrors) Len() int {
	if es == nil {
		return 0
	}
	return len(es.Errors)
}

// Error implements the error interface.
func (es *ValidationErrors) Error() string {
	if !es.HasErrors() {
		return ""
	}
	parts := make([]string, len(es.Errors))
	for i, e := range es.Errors {
		parts[i] = e.Error()
	}
	return strings.Join(parts, "; ")
}

// Strings renders every error with the given locale and format template.
func (es *ValidationErrors) Strings(locale, format string) []string {
	if !es.HasErrors() {
		return nil
	}
	out := make([]string, 0, len(es.Errors))
	for _, e := range es.Errors {
		out = append(out, e.Format(locale, format))
	}
	return out
}

// Messages returns just the localized messages.
func (es *ValidationErrors) Messages(locale string) []string {
	if !es.HasErrors() {
		return nil
	}
	out := make([]string, 0, len(es.Errors))
	for _, e := range es.Errors {
		out = append(out, e.Message(locale))
	}
	return out
}

// ForTarget returns the subset of errors whose Target equals target.
func (es *ValidationErrors) ForTarget(target string) []*ValidationError {
	if es == nil {
		return nil
	}
	var out []*ValidationError
	for _, e := range es.Errors {
		if e.Target == target {
			out = append(out, e)
		}
	}
	return out
}

// SchemaError is returned when a schema references a rule, operator, or
// condition that is not registered, or is otherwise malformed. It is distinct
// from ValidationErrors, which reports problems with the data.
type SchemaError struct {
	Problems []string
}

func (e *SchemaError) Error() string {
	return "invalid schema: " + strings.Join(e.Problems, "; ")
}
