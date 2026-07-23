package schematics

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Schematics is the entry point: it holds a schema, the registries of
// validators/operators/conditions (pre-loaded with the built-ins), and
// configuration. Create one with New, load a schema, then call Validate or
// Operate.
type Schematics struct {
	schema     Schema
	rules      map[string]Rule
	operators  map[string]Operator
	conditions map[string]Condition

	separator  string
	arrayIDKey string
	locale     string
	db         map[string]any
	logger     *slog.Logger

	// user-set overrides win over values found in a loaded schema file.
	sepSet    bool
	localeSet bool
	arrIDSet  bool

	checked  bool
	checkErr error
}

// Option configures a Schematics at construction time.
type Option func(*Schematics)

// WithSeparator sets the key separator used when flattening documents.
func WithSeparator(sep string) Option {
	return func(s *Schematics) {
		if sep != "" {
			s.separator = sep
			s.sepSet = true
		}
	}
}

// WithLocale sets the default locale for error messages.
func WithLocale(locale string) Option {
	return func(s *Schematics) {
		if locale != "" {
			s.locale = locale
			s.localeSet = true
		}
	}
}

// WithArrayIDKey sets the flattened key whose value identifies each row when
// validating an array of objects.
func WithArrayIDKey(key string) Option {
	return func(s *Schematics) {
		s.arrayIDKey = key
		s.arrIDSet = true
	}
}

// WithDB seeds the shared DB that is passed to every rule.
func WithDB(db map[string]any) Option {
	return func(s *Schematics) {
		for k, v := range db {
			s.db[k] = v
		}
	}
}

// WithLogger attaches a slog.Logger. By default logs are discarded.
func WithLogger(l *slog.Logger) Option {
	return func(s *Schematics) {
		if l != nil {
			s.logger = l
		}
	}
}

// New creates a Schematics with every built-in validator, operator, and
// condition registered, then applies the given options.
func New(opts ...Option) *Schematics {
	s := &Schematics{
		rules:      map[string]Rule{},
		operators:  map[string]Operator{},
		conditions: map[string]Condition{},
		separator:  defaultSeparator,
		locale:     defaultLocale,
		db:         map[string]any{},
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	registerBuiltinRules(s)
	registerBuiltinOperators(s)
	registerBuiltinConditions(s)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ImportSchema is a convenience constructor for callers who start from raw
// schema bytes (a config file already read into memory, a network response,
// an embedded fixture): it combines New and LoadBytes into one call.
//
//	s, err := schematics.ImportSchema(schemaBytes)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Any options are applied the same way they are for New, before the schema
// is loaded.
func ImportSchema(b []byte, opts ...Option) (*Schematics, error) {
	s := New(opts...)
	if err := s.LoadBytes(b); err != nil {
		return nil, err
	}
	return s, nil
}

// SetSchema installs a schema value directly.
func (s *Schematics) SetSchema(schema Schema) *Schematics {
	s.schema = schema
	s.applySchemaConfig()
	s.checked = false
	return s
}

// Schema returns the currently loaded schema.
func (s *Schematics) Schema() Schema { return s.schema }

// LoadBytes parses a JSON schema from b.
func (s *Schematics) LoadBytes(b []byte) error {
	var schema Schema
	if err := json.Unmarshal(b, &schema); err != nil {
		s.logger.Error("failed to parse schema", "err", err)
		return fmt.Errorf("parse schema: %w", err)
	}
	s.SetSchema(schema)
	return nil
}

// LoadFile reads and parses a JSON schema file.
func (s *Schematics) LoadFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		s.logger.Error("failed to read schema file", "path", path, "err", err)
		return fmt.Errorf("read schema file: %w", err)
	}
	return s.LoadBytes(b)
}

// LoadMap parses a schema from an in-memory value (a map or struct) by
// round-tripping it through JSON.
func (s *Schematics) LoadMap(m any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal schema map: %w", err)
	}
	return s.LoadBytes(b)
}

func (s *Schematics) applySchemaConfig() {
	if !s.sepSet && s.schema.Separator != "" {
		s.separator = s.schema.Separator
	}
	if !s.localeSet && s.schema.Locale != "" {
		s.locale = s.schema.Locale
	}
	if !s.arrIDSet && s.schema.ArrayIDKey != "" {
		s.arrayIDKey = s.schema.ArrayIDKey
	}
	for k, v := range s.schema.DB {
		if _, exists := s.db[k]; !exists {
			s.db[k] = v
		}
	}
}

// Check verifies that every rule, operator, and condition referenced by the
// schema is registered and that fields are well formed. It returns a
// *SchemaError describing all problems, or nil.
func (s *Schematics) Check() error {
	var problems []string
	seen := map[string]bool{}
	for i, f := range s.schema.Fields {
		if f.Target == "" {
			problems = append(problems, fmt.Sprintf("field #%d has an empty target", i))
		}
		if seen[f.Target] && f.Target != "" {
			problems = append(problems, fmt.Sprintf("duplicate target %q", f.Target))
		}
		seen[f.Target] = true
		for _, r := range f.Validate {
			if _, ok := s.rules[r.Rule]; !ok {
				problems = append(problems, fmt.Sprintf("target %q references unknown rule %q", f.Target, r.Rule))
			}
		}
		for _, o := range f.Operate {
			if _, ok := s.operators[o.Op]; !ok {
				problems = append(problems, fmt.Sprintf("target %q references unknown operator %q", f.Target, o.Op))
			}
		}
		for _, w := range f.When {
			if _, ok := s.conditions[w.Condition]; !ok {
				problems = append(problems, fmt.Sprintf("target %q references unknown condition %q", f.Target, w.Condition))
			}
		}
	}
	if len(problems) > 0 {
		return &SchemaError{Problems: problems}
	}
	return nil
}

func (s *Schematics) ensureChecked() error {
	if !s.checked {
		s.checkErr = s.Check()
		s.checked = true
	}
	return s.checkErr
}
