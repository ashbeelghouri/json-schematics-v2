// Package schematics validates and transforms arbitrary JSON documents against
// a declarative, data-driven schema.
//
// The model is simple: a document is flattened into dotted keys (for example
// user.profile.name or tags.0), each schema field targets one or more of those
// keys (literally, with a * wildcard, or with a full regular expression), and
// every matched value is run through a chain of validators and operators.
// Validators report problems as typed ValidationError values; operators
// transform the value in place.
//
// The package has no third-party dependencies.
package schematics

// Version is the schema/library generation this package implements.
const Version = "2.0"

const (
	defaultSeparator = "."
	defaultLocale    = "en"
)
