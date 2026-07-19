package schematics

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// APISchema describes request validation for a set of HTTP endpoints. Each
// endpoint validates its headers, query parameters, and JSON body using the
// same Field model as the core schema.
//
// Example (JSON):
//
//	{
//	  "version": "2.0",
//	  "global": { "headers": [ { "target": "authorization", "required": true } ] },
//	  "endpoints": [
//	    {
//	      "path": "/users/:id",
//	      "method": "POST",
//	      "query":  [ { "target": "verbose", "validate": [ { "rule": "inOptions", "args": { "options": ["true","false"] } } ] } ],
//	      "body":   [ { "target": "email", "required": true, "validate": [ { "rule": "email" } ] } ]
//	    }
//	  ]
//	}
type APISchema struct {
	Version   string        `json:"version,omitempty"`
	Separator string        `json:"separator,omitempty"`
	Global    APIGlobal     `json:"global,omitempty"`
	Endpoints []APIEndpoint `json:"endpoints"`
}

// APIGlobal holds fields applied to every endpoint (currently headers).
type APIGlobal struct {
	Headers []Field `json:"headers,omitempty"`
}

// APIEndpoint validates one method+path combination.
type APIEndpoint struct {
	Path    string  `json:"path"`
	Method  string  `json:"method"`
	Headers []Field `json:"headers,omitempty"`
	Query   []Field `json:"query,omitempty"`
	Body    []Field `json:"body,omitempty"`
}

// API validates *http.Request values against an APISchema. It reuses a base
// Schematics for its registries and options, so custom rules registered there
// are available to request validation too.
type API struct {
	schema APISchema
	base   *Schematics
}

// NewAPI creates an API validator with the given options (shared with the
// underlying engine).
func NewAPI(opts ...Option) *API {
	return &API{base: New(opts...)}
}

// Base exposes the underlying Schematics so callers can register custom rules,
// operators, and conditions used by the request schema.
func (a *API) Base() *Schematics { return a.base }

// LoadBytes parses an API schema from JSON.
func (a *API) LoadBytes(b []byte) error {
	var schema APISchema
	if err := json.Unmarshal(b, &schema); err != nil {
		return err
	}
	a.schema = schema
	return nil
}

// LoadFile reads and parses an API schema file.
func (a *API) LoadFile(path string) error {
	b, err := readFile(path)
	if err != nil {
		return err
	}
	return a.LoadBytes(b)
}

// ValidateRequest validates r against the endpoint matching its method and
// path. It returns nil, a *ValidationErrors, or a *SchemaError. If no endpoint
// matches, it returns a *ValidationErrors describing the mismatch.
func (a *API) ValidateRequest(r *http.Request) error {
	ep, ok := a.match(r.Method, r.URL.Path)
	if !ok {
		errs := &ValidationErrors{}
		e := newError(r.URL.Path, "route", r.Method+" "+r.URL.Path)
		e.fallback = "no matching endpoint"
		errs.Add(e)
		return errs
	}

	errs := &ValidationErrors{}

	// headers (global + endpoint)
	headerFields := append(append([]Field{}, a.schema.Global.Headers...), ep.Headers...)
	if len(headerFields) > 0 {
		headers := map[string]any{}
		for k := range r.Header {
			headers[strings.ToLower(k)] = r.Header.Get(k)
		}
		a.validateSection(headerFields, headers, errs)
	}

	// query
	if len(ep.Query) > 0 {
		query := map[string]any{}
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				query[k] = v[0]
			}
		}
		a.validateSection(ep.Query, query, errs)
	}

	// body
	if len(ep.Body) > 0 {
		body := map[string]any{}
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(strings.NewReader(string(b)))
			if len(b) > 0 {
				_ = json.Unmarshal(b, &body)
			}
		}
		a.validateSection(ep.Body, body, errs)
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (a *API) validateSection(fields []Field, data map[string]any, errs *ValidationErrors) {
	sub := a.base.withFields(fields)
	if err := sub.Validate(data); err != nil {
		if ve, ok := err.(*ValidationErrors); ok {
			for _, e := range ve.Errors {
				errs.Add(e)
			}
			return
		}
		// schema error: surface as a single entry
		e := newError("$schema", "schema", nil)
		e.fallback = err.Error()
		errs.Add(e)
	}
}

// match finds the endpoint whose method and path pattern match. Patterns use
// ":name" for a single path segment and "*" for a trailing wildcard.
func (a *API) match(method, path string) (APIEndpoint, bool) {
	for _, ep := range a.schema.Endpoints {
		if !strings.EqualFold(ep.Method, method) {
			continue
		}
		if pathMatches(ep.Path, path) {
			return ep, true
		}
	}
	return APIEndpoint{}, false
}

func pathMatches(pattern, path string) bool {
	if pattern == path {
		return true
	}
	pp := strings.Split(strings.Trim(pattern, "/"), "/")
	sp := strings.Split(strings.Trim(path, "/"), "/")
	for i, seg := range pp {
		if seg == "*" {
			return true // trailing wildcard matches the rest
		}
		if i >= len(sp) {
			return false
		}
		if strings.HasPrefix(seg, ":") {
			continue // named param matches any single segment
		}
		if seg != sp[i] {
			return false
		}
	}
	return len(pp) == len(sp)
}

// withFields returns a shallow clone of s bound to a new field set, sharing the
// registries and configuration.
func (s *Schematics) withFields(fields []Field) *Schematics {
	clone := *s
	clone.schema = Schema{Version: s.schema.Version, Fields: fields}
	clone.checked = false
	clone.checkErr = nil
	return &clone
}
